package oauth

import (
	"sync"
	"time"
)

// StateData holds transient OIDC authorization state
type StateData struct {
	CodeVerifier string
	Nonce        string
	ExpiresAt    time.Time
}

// StateStore abstracts storage for OIDC state between login and callback
type StateStore interface {
	Save(state string, data StateData)
	GetAndDelete(state string) (StateData, bool)
}

// InMemoryStateStore is a simple TTL map suitable for single-instance dev
type InMemoryStateStore struct {
	mu   sync.Mutex
	data map[string]StateData
}

func NewInMemoryStateStore() *InMemoryStateStore {
	s := &InMemoryStateStore{data: make(map[string]StateData)}
	go s.gc()
	return s
}

func (s *InMemoryStateStore) Save(state string, data StateData) {
	s.mu.Lock()
	s.data[state] = data
	s.mu.Unlock()
}

func (s *InMemoryStateStore) GetAndDelete(state string) (StateData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.data[state]
	if ok {
		delete(s.data, state)
		if time.Now().After(d.ExpiresAt) {
			return StateData{}, false
		}
	}
	return d, ok
}

func (s *InMemoryStateStore) gc() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, v := range s.data {
			if now.After(v.ExpiresAt) {
				delete(s.data, k)
			}
		}
		s.mu.Unlock()
	}
}
