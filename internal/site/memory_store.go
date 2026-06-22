package site

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu          sync.Mutex
	sites       map[string]Site
	credentials map[string]TrackingCredential
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sites:       make(map[string]Site),
		credentials: make(map[string]TrackingCredential),
	}
}

func (s *MemoryStore) CreateSite(_ context.Context, site Site, credential TrackingCredential) (Site, TrackingCredential, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sites[site.ID] = site
	s.credentials[credential.PublicKey] = credential
	return site, credential, nil
}
