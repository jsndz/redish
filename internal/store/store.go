package store

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/jsndz/redish/internal/client"
)

type kind string

const (
	list kind = "list"
	str  kind = "string"
)

type Store struct {
	mu          sync.RWMutex
	data        map[string]*entry
	watchedkeys map[string]map[*client.Client]bool
}

type entry struct {
	kind      kind
	value     string
	listValue []string
	timer     *time.Timer
}

func New() *Store {
	return &Store{
		data:        make(map[string]*entry),
		watchedkeys: map[string]map[*client.Client]bool{},
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
	s.touchWatcher(key)
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
	s.touchWatcher(key)

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
	s.touchWatcher(key)
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
		s.touchWatcher(key)
		return 1, nil
	}
	intVal, err := strconv.Atoi(val.value)
	if err != nil {
		return 0, fmt.Errorf("value is not a number")
	}
	intVal++
	s.data[key].value = strconv.Itoa(intVal)
	s.touchWatcher(key)
	return intVal, nil
}

func (s *Store) AddWatcher(key string, c *client.Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watchedkeys[key] == nil {
		s.watchedkeys[key] = make(map[*client.Client]bool)
	}
	s.watchedkeys[key][c] = true
	if c.WatchedKeys == nil {
		c.WatchedKeys = map[string]bool{}
	}
	c.WatchedKeys[key] = true
	return nil
}

// if the key changes we call touch watchers immediately
// this will mark all watchers dirty
// if the client is dirty EXEC wont run
func (s *Store) touchWatcher(key string) {
	watchers, ok := s.watchedkeys[key]
	if !ok {
		return
	}
	for c := range watchers {
		c.DirtyCAS = true
	}
}

func (s *Store) RemoveAllWatchers(c *client.Client) {
	for key, watchers := range s.watchedkeys {
		if watchers[c] {
			delete(watchers, c)
			if len(watchers) == 0 {
				delete(s.watchedkeys, key)
			}
		}
	}
	c.WatchedKeys = nil
	c.DirtyCAS = false
}
