package store

import (
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]*entry
}

type entry struct {
	value string
	timer *time.Timer
}

func New() *Store {
	return &Store{
		data: make(map[string]*entry),
	}
}

func (s *Store) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.data[key]; ok && existing.timer != nil {
		existing.timer.Stop()
	}

	item := &entry{value: value}
	s.data[key] = item

	if ttl > 0 {
		k := key
		item.timer = time.AfterFunc(ttl, func() {
			s.Delete(k)
		})
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.data[key]
	if !ok {
		return "", false
	}

	return item.value, true
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, ok := s.data[key]; ok && item.timer != nil {
		item.timer.Stop()
	}
	delete(s.data, key)
}
