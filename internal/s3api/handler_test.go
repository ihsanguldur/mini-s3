package s3api

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ihsanguldur/mini-s3/internal/metadata"
)

func newServer() *http.ServeMux {
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(metadata.New()))
	return mux
}

func do(mux *http.ServeMux, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

// assertError checks both the HTTP status and the S3 error code in the body.
func assertError(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, status, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/xml" {
		t.Fatalf("Content-Type = %q, want application/xml", ct)
	}
	var resp errorResponse
	if err := xml.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error body is not valid XML: %v (body: %s)", err, rec.Body)
	}
	if resp.Code != code {
		t.Fatalf("error code = %q, want %q", resp.Code, code)
	}
}

func listBuckets(t *testing.T, mux *http.ServeMux) listAllMyBucketsResult {
	t.Helper()
	rec := do(mux, "GET", "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("ListBuckets status = %d", rec.Code)
	}
	var resp listAllMyBucketsResult
	if err := xml.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("ListBuckets body is not valid XML: %v", err)
	}
	return resp
}

func TestCreateBucket(t *testing.T) {
	mux := newServer()
	rec := do(mux, "PUT", "/photos")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/photos" {
		t.Fatalf("Location = %q, want /photos", loc)
	}

	assertError(t, do(mux, "PUT", "/photos"), http.StatusConflict, "BucketAlreadyOwnedByYou")
	assertError(t, do(mux, "PUT", "/Photos"), http.StatusBadRequest, "InvalidBucketName")
	assertError(t, do(mux, "PUT", "/ab"), http.StatusBadRequest, "InvalidBucketName")
	assertError(t, do(mux, "PUT", "/a%2Fb"), http.StatusBadRequest, "InvalidBucketName")
}

func TestHeadBucket(t *testing.T) {
	mux := newServer()
	if rec := do(mux, "HEAD", "/photos"); rec.Code != http.StatusNotFound {
		t.Fatalf("HEAD missing bucket status = %d, want 404", rec.Code)
	}
	do(mux, "PUT", "/photos")
	if rec := do(mux, "HEAD", "/photos"); rec.Code != http.StatusOK {
		t.Fatalf("HEAD status = %d, want 200", rec.Code)
	}
}

func TestDeleteBucket(t *testing.T) {
	mux := newServer()
	assertError(t, do(mux, "DELETE", "/photos"), http.StatusNotFound, "NoSuchBucket")

	do(mux, "PUT", "/photos")
	if rec := do(mux, "DELETE", "/photos"); rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want 204", rec.Code)
	}
	if rec := do(mux, "HEAD", "/photos"); rec.Code != http.StatusNotFound {
		t.Fatalf("HEAD after delete status = %d, want 404", rec.Code)
	}
}

func TestTrailingSlashRoutes(t *testing.T) {
	mux := newServer()
	if rec := do(mux, "PUT", "/photos/"); rec.Code != http.StatusOK {
		t.Fatalf("PUT /photos/ status = %d, want 200", rec.Code)
	}
	if rec := do(mux, "HEAD", "/photos/"); rec.Code != http.StatusOK {
		t.Fatalf("HEAD /photos/ status = %d, want 200", rec.Code)
	}
	if rec := do(mux, "DELETE", "/photos/"); rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /photos/ status = %d, want 204", rec.Code)
	}
}

func TestListBuckets(t *testing.T) {
	mux := newServer()

	rec := do(mux, "GET", "/")
	if !strings.Contains(rec.Body.String(), "<Buckets></Buckets>") {
		t.Fatalf("empty list must still contain <Buckets></Buckets>, got %s", rec.Body)
	}

	for _, name := range []string{"zeta", "alpha", "mid"} {
		do(mux, "PUT", "/"+name)
	}
	resp := listBuckets(t, mux)
	if resp.Xmlns != s3Namespace {
		t.Fatalf("xmlns = %q, want %q", resp.Xmlns, s3Namespace)
	}
	if resp.Owner != defaultOwner {
		t.Fatalf("Owner = %+v, want %+v", resp.Owner, defaultOwner)
	}
	want := []string{"alpha", "mid", "zeta"}
	if len(resp.Buckets.Bucket) != len(want) {
		t.Fatalf("got %d buckets, want %d", len(resp.Buckets.Bucket), len(want))
	}
	for i, b := range resp.Buckets.Bucket {
		if b.Name != want[i] {
			t.Fatalf("bucket[%d] = %s, want %s", i, b.Name, want[i])
		}
		if _, err := time.Parse(timeFormat, b.CreationDate); err != nil {
			t.Fatalf("CreationDate %q is not in S3 format: %v", b.CreationDate, err)
		}
	}
}

func TestSubresourceQueriesNotImplemented(t *testing.T) {
	mux := newServer()
	assertError(t, do(mux, "PUT", "/photos?versioning"), http.StatusNotImplemented, "NotImplemented")
	if rec := do(mux, "HEAD", "/photos"); rec.Code != http.StatusNotFound {
		t.Fatal("PUT with a subresource query must not create the bucket")
	}

	do(mux, "PUT", "/photos")
	assertError(t, do(mux, "DELETE", "/photos?cors"), http.StatusNotImplemented, "NotImplemented")
	if rec := do(mux, "HEAD", "/photos"); rec.Code != http.StatusOK {
		t.Fatal("DELETE with a subresource query must not delete the bucket")
	}
}

func TestUnknownRoutesNotImplemented(t *testing.T) {
	mux := newServer()
	for _, req := range []struct{ method, path string }{
		{"GET", "/photos"},
		{"PUT", "/photos/cat.jpg"},
		{"POST", "/"},
	} {
		assertError(t, do(mux, req.method, req.path), http.StatusNotImplemented, "NotImplemented")
	}
}
