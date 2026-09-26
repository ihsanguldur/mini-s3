package s3api

import (
	"net/http"

	"github.com/ihsanguldur/mini-s3/internal/metadata"
)

type Handler struct {
	meta *metadata.Store
}

func NewHandler(meta *metadata.Store) *Handler {
	return &Handler{meta: meta}
}

func (h *Handler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	resp := listAllMyBucketsResult{Xmlns: s3Namespace, Owner: defaultOwner}
	for _, b := range h.meta.ListBuckets() {
		resp.Buckets.Bucket = append(resp.Buckets.Bucket, bucketEntry{
			Name:         b.Name,
			CreationDate: b.CreatedAt.Format(timeFormat),
		})
	}
	writeXML(w, http.StatusOK, resp)
}

func (h *Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		writeError(w, r, errNotImplemented)
		return
	}
	name := r.PathValue("bucket")
	if !validBucketName(name) {
		writeError(w, r, errInvalidBucketName)
		return
	}
	if err := h.meta.CreateBucket(name); err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Location", "/"+name)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HeadBucket(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		writeError(w, r, errNotImplemented)
		return
	}
	if _, err := h.meta.Bucket(r.PathValue("bucket")); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		writeError(w, r, errNotImplemented)
		return
	}
	if err := h.meta.DeleteBucket(r.PathValue("bucket")); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) NotImplemented(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, errNotImplemented)
}
