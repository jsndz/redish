package store

import (
	"fmt"
	"sync"
	"time"
)

type kind string

const (
	list kind = "list"
	str  kind = "string"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]*entry
}

type entry struct {
	kind      kind
	value     string
	listValue []string
	timer     *time.Timer
}

func New() *Store {
	return &Store{
		data: make(map[string]*entry),
	}
}

func (s *Store) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.data[key]
	if ok {
		if existing.kind != str {
			return
		}
		if existing.timer != nil {
			existing.timer.Stop()
		}

	}

	item := &entry{kind: str, value: value}
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

func (s *Store) Rpush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[key]
	if ok {
		if item.kind != list {
			return 0, fmt.Errorf("wrong type")
		}
	} else {
		item = &entry{kind: list, listValue: []string{}}
		s.data[key] = item
	}
	item.listValue = append(item.listValue, values...)
	return len(item.listValue), nil
}
