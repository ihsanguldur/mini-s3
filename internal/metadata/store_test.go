package metadata

import (
	"errors"
	"sync"
	"testing"
)

func TestCreateAndGetBucket(t *testing.T) {
	s := New()
	if err := s.CreateBucket("photos"); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	b, err := s.Bucket("photos")
	if err != nil {
		t.Fatalf("Bucket: %v", err)
	}
	if b.Name != "photos" || b.CreatedAt.IsZero() {
		t.Fatalf("Bucket = %+v, want name photos and a creation time", b)
	}
	if b.CreatedAt.Location().String() != "UTC" {
		t.Fatalf("CreatedAt location = %s, want UTC", b.CreatedAt.Location())
	}
}

func TestCreateBucketTwice(t *testing.T) {
	s := New()
	if err := s.CreateBucket("photos"); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	first, _ := s.Bucket("photos")

	if err := s.CreateBucket("photos"); !errors.Is(err, ErrBucketExists) {
		t.Fatalf("second CreateBucket err = %v, want ErrBucketExists", err)
	}
	if again, _ := s.Bucket("photos"); !again.CreatedAt.Equal(first.CreatedAt) {
		t.Fatal("second CreateBucket overwrote the original bucket")
	}
}

func TestMissingBucket(t *testing.T) {
	s := New()
	if _, err := s.Bucket("nope"); !errors.Is(err, ErrBucketNotFound) {
		t.Fatalf("Bucket err = %v, want ErrBucketNotFound", err)
	}
	if err := s.DeleteBucket("nope"); !errors.Is(err, ErrBucketNotFound) {
		t.Fatalf("DeleteBucket err = %v, want ErrBucketNotFound", err)
	}
}

func TestDeleteBucket(t *testing.T) {
	s := New()
	if err := s.CreateBucket("photos"); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	if err := s.DeleteBucket("photos"); err != nil {
		t.Fatalf("DeleteBucket: %v", err)
	}
	if _, err := s.Bucket("photos"); !errors.Is(err, ErrBucketNotFound) {
		t.Fatalf("Bucket after delete err = %v, want ErrBucketNotFound", err)
	}
	if err := s.CreateBucket("photos"); err != nil {
		t.Fatalf("re-create after delete: %v", err)
	}
}

func TestListBucketsSorted(t *testing.T) {
	s := New()
	if got := s.ListBuckets(); len(got) != 0 {
		t.Fatalf("ListBuckets on empty store = %v, want empty", got)
	}
	for _, name := range []string{"zeta", "alpha", "mid", "alpha2"} {
		if err := s.CreateBucket(name); err != nil {
			t.Fatalf("CreateBucket(%s): %v", name, err)
		}
	}
	want := []string{"alpha", "alpha2", "mid", "zeta"}
	got := s.ListBuckets()
	if len(got) != len(want) {
		t.Fatalf("ListBuckets returned %d buckets, want %d", len(got), len(want))
	}
	for i, b := range got {
		if b.Name != want[i] {
			t.Fatalf("ListBuckets[%d] = %s, want %s", i, b.Name, want[i])
		}
	}
}

func TestConcurrentCreateSameBucket(t *testing.T) {
	s := New()
	var wg sync.WaitGroup
	errs := make(chan error, 50)
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- s.CreateBucket("photos")
		}()
	}
	wg.Wait()
	close(errs)

	created := 0
	for err := range errs {
		switch {
		case err == nil:
			created++
		case !errors.Is(err, ErrBucketExists):
			t.Fatalf("CreateBucket: %v", err)
		}
	}
	if created != 1 {
		t.Fatalf("%d concurrent CreateBucket calls succeeded, want exactly 1", created)
	}
}
