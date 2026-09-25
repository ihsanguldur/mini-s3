package chunkstore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrInvalidID    = errors.New("invalid chunk ID")
	ErrHashMismatch = errors.New("chunk content does not match its id")
	ErrNotFound     = errors.New("chunk not found")
)

type Store struct {
	dir    string
	tmpDir string
}

func New(dir string) (*Store, error) {
	tmpDir := filepath.Join(dir, "tmp")
	if err := os.RemoveAll(tmpDir); err != nil {
		return nil, fmt.Errorf("clean tmp dir: %w", err)
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, fmt.Errorf("create tmp dir: %w", err)
	}
	return &Store{dir: dir, tmpDir: tmpDir}, nil
}

func (s *Store) Put(id string, r io.Reader) error {
	if !validID(id) {
		return ErrInvalidID
	}
	final := s.path(id)
	if _, err := os.Stat(final); err == nil {
		return nil
	}

	tmp, err := os.CreateTemp(s.tmpDir, id+"-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, h), r); err != nil {
		return fmt.Errorf("write chunk: %w", err)
	}
	if hex.EncodeToString(h.Sum(nil)) != id {
		return ErrHashMismatch
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	dir := filepath.Dir(final)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), final); err != nil {
		return err
	}
	return syncDir(dir)
}

func (s *Store) Get(id string) (io.ReadCloser, int64, error) {
	if !validID(id) {
		return nil, 0, ErrInvalidID
	}
	f, err := os.Open(s.path(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, ErrNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}

func (s *Store) Size(id string) (int64, error) {
	if !validID(id) {
		return 0, ErrInvalidID
	}
	info, err := os.Stat(s.path(id))
	if errors.Is(err, os.ErrNotExist) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (s *Store) Delete(id string) error {
	if !validID(id) {
		return ErrInvalidID
	}
	err := os.Remove(s.path(id))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Store) path(id string) string {
	return filepath.Join(s.dir, id[0:2], id[2:4], id)
}

func validID(id string) bool {
	if len(id) != sha256.Size*2 {
		return false
	}
	for _, c := range id {
		if !('0' <= c && c <= '9' || 'a' <= c && c <= 'f') {
			return false
		}
	}
	return true
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
