package session

import (
	"errors"
	"sync"
	"time"
)

// MemStore 内存会话存储（默认，零外部依赖）
type MemStore struct {
	mu    sync.RWMutex
	items map[string]*memEntry
}

type memEntry struct {
	value string
	exp   time.Time
}

func NewMemStore() *MemStore {
	s := &MemStore{items: make(map[string]*memEntry)}
	go s.cleanup()
	return s
}

func (s *MemStore) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[key]
	if !ok || (!e.exp.IsZero() && time.Now().After(e.exp)) {
		return "", errors.New("not found")
	}
	return e.value, nil
}

func (s *MemStore) Set(key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	s.items[key] = &memEntry{value: value, exp: exp}
	return nil
}

func (s *MemStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func (s *MemStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, v := range s.items {
			if !v.exp.IsZero() && now.After(v.exp) {
				delete(s.items, k)
			}
		}
		s.mu.Unlock()
	}
}
