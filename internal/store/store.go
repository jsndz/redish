package store

import (
	"fmt"
	"strconv"
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

func (s *Store) Lrange(key string, start, end int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("key doesn't exist")
	}
	if item.kind != list {
		return nil, fmt.Errorf("wrong type")
	}
	length := len(item.listValue)

	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}

	if end >= length {
		end = length - 1
	}

	if end < 0 {
		return []string{}, nil
	}

	if start < 0 {
		start = 0
	}

	if start >= length || start > end {
		return []string{}, nil
	}
	values := []string{}
	for i := start; i <= end; i++ {
		values = append(values, item.listValue[i])
	}
	return values, nil
}

func (s *Store) Incr(key string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.data[key]
	if !ok {
		s.data[key] = &entry{kind: str, value: "1"}
		return 1, nil
	}
	intVal, err := strconv.Atoi(val.value)
	if err != nil {
		return 0, fmt.Errorf("value is not a number")
	}
	intVal++
	s.data[key].value = strconv.Itoa(intVal)
	return intVal, nil
}
