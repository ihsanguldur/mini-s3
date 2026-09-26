package metadata

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"time"
)

var (
	ErrBucketExists   = errors.New("bucket already exists")
	ErrBucketNotFound = errors.New("bucket not found")
)

type Bucket struct {
	Name      string
	CreatedAt time.Time
}

type Store struct {
	mu      sync.RWMutex
	buckets map[string]Bucket
}

func New() *Store {
	return &Store{
		buckets: make(map[string]Bucket),
	}
}

func (s *Store) CreateBucket(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[name]; ok {
		return ErrBucketExists
	}
	s.buckets[name] = Bucket{Name: name, CreatedAt: time.Now().UTC()}
	return nil
}

func (s *Store) Bucket(name string) (Bucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.buckets[name]
	if !ok {
		return Bucket{}, ErrBucketNotFound
	}
	return b, nil
}

func (s *Store) DeleteBucket(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[name]; !ok {
		return ErrBucketNotFound
	}
	delete(s.buckets, name)
	return nil
}

func (s *Store) ListBuckets() []Bucket {
	s.mu.RLock()
	list := make([]Bucket, 0, len(s.buckets))
	for _, b := range s.buckets {
		list = append(list, b)
	}
	s.mu.RUnlock()

	slices.SortFunc(list, func(a, b Bucket) int {
		return strings.Compare(a.Name, b.Name)
	})
	return list
}
