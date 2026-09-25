package chunkstore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestPutGet(t *testing.T) {
	s, dir := newStore(t)
	id := idOf("hello")

	if err := s.Put(id, strings.NewReader("hello")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if got := readChunk(t, s, id); got != "hello" {
		t.Fatalf("Get = %q, want %q", got, "hello")
	}
	if _, err := os.Stat(filepath.Join(dir, id[0:2], id[2:4], id)); err != nil {
		t.Fatalf("chunk not at fan-out path: %v", err)
	}
}

func TestPutHashMismatch(t *testing.T) {
	s, dir := newStore(t)
	id := idOf("hello")

	if err := s.Put(id, strings.NewReader("hellX")); !errors.Is(err, ErrHashMismatch) {
		t.Fatalf("Put err = %v, want ErrHashMismatch", err)
	}
	if _, _, err := s.Get(id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get err = %v, want ErrNotFound", err)
	}
	assertTmpEmpty(t, dir)
}

func TestPutReadErrorLeavesNothing(t *testing.T) {
	s, dir := newStore(t)
	id := idOf("hello")
	body := io.MultiReader(strings.NewReader("hel"), errReader{})

	if err := s.Put(id, body); err == nil {
		t.Fatal("Put succeeded on a broken body")
	}
	if _, _, err := s.Get(id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get err = %v, want ErrNotFound", err)
	}
	assertTmpEmpty(t, dir)
}

func TestPutExistingIsNoop(t *testing.T) {
	s, _ := newStore(t)
	id := idOf("hello")
	if err := s.Put(id, strings.NewReader("hello")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// The body must not even be read when the chunk already exists.
	if err := s.Put(id, errReader{}); err != nil {
		t.Fatalf("second Put: %v", err)
	}
	if got := readChunk(t, s, id); got != "hello" {
		t.Fatalf("Get = %q, want %q", got, "hello")
	}
}

func TestInvalidID(t *testing.T) {
	s, _ := newStore(t)
	valid := idOf("hello")
	ids := map[string]string{
		"empty":     "",
		"short":     "abc",
		"uppercase": strings.ToUpper(valid),
		"non-hex":   strings.Repeat("g", 64),
		"traversal": "../../../../etc/passwd" + strings.Repeat("0", 42),
	}
	for name, id := range ids {
		t.Run(name, func(t *testing.T) {
			if err := s.Put(id, strings.NewReader("hello")); !errors.Is(err, ErrInvalidID) {
				t.Errorf("Put err = %v", err)
			}
			if _, _, err := s.Get(id); !errors.Is(err, ErrInvalidID) {
				t.Errorf("Get err = %v", err)
			}
			if _, err := s.Size(id); !errors.Is(err, ErrInvalidID) {
				t.Errorf("Size err = %v", err)
			}
			if err := s.Delete(id); !errors.Is(err, ErrInvalidID) {
				t.Errorf("Delete err = %v", err)
			}
		})
	}
}

func TestSize(t *testing.T) {
	s, _ := newStore(t)
	id := idOf("hello")
	if _, err := s.Size(id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Size err = %v, want ErrNotFound", err)
	}
	if err := s.Put(id, strings.NewReader("hello")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if size, err := s.Size(id); err != nil || size != 5 {
		t.Fatalf("Size = %d, %v; want 5, nil", size, err)
	}
}

func TestDelete(t *testing.T) {
	s, _ := newStore(t)
	id := idOf("hello")
	if err := s.Put(id, strings.NewReader("hello")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	if err := s.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, _, err := s.Get(id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after Delete err = %v, want ErrNotFound", err)
	}
	if err := s.Delete(id); err != nil {
		t.Fatalf("second Delete: %v", err)
	}
}

func TestReopenKeepsChunksAndCleansTmp(t *testing.T) {
	s, dir := newStore(t)
	id := idOf("hello")
	if err := s.Put(id, strings.NewReader("hello")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	// Simulate a write that crashed before its rename.
	if err := os.WriteFile(filepath.Join(dir, "tmp", "leftover"), []byte("hel"), 0o644); err != nil {
		t.Fatal(err)
	}

	reopened, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := readChunk(t, reopened, id); got != "hello" {
		t.Fatalf("Get after reopen = %q, want %q", got, "hello")
	}
	assertTmpEmpty(t, dir)
}

func TestConcurrentPutSameChunk(t *testing.T) {
	s, dir := newStore(t)
	data := strings.Repeat("x", 1<<20)
	id := idOf(data)

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- s.Put(id, strings.NewReader(data))
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Put: %v", err)
		}
	}
	if got := readChunk(t, s, id); got != data {
		t.Fatal("content corrupted by concurrent Puts")
	}
	assertTmpEmpty(t, dir)
}

func newStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, dir
}

func idOf(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func readChunk(t *testing.T, s *Store, id string) string {
	t.Helper()
	rc, size, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read chunk: %v", err)
	}
	if int64(len(b)) != size {
		t.Fatalf("Get size = %d, but read %d bytes", size, len(b))
	}
	return string(b)
}

func assertTmpEmpty(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, "tmp"))
	if err != nil {
		t.Fatalf("read tmp dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("tmp dir has %d leftover files", len(entries))
	}
}

// errReader fails every read, standing in for a client that disconnects.
type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("connection reset")
}
