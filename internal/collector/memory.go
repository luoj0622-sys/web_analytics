package collector

import (
	"context"
	"sync"
	"time"

	"webanalytics/internal/domain"
	"webanalytics/internal/queue"
)

type MemoryCredentials struct {
	mu          sync.RWMutex
	credentials map[string]Credential
}

func NewMemoryCredentials() *MemoryCredentials {
	return &MemoryCredentials{credentials: make(map[string]Credential)}
}

func (s *MemoryCredentials) Add(publicKey string, credential Credential) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[publicKey] = credential
}

func (s *MemoryCredentials) Validate(_ context.Context, siteID, publicKey string) (Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	credential, ok := s.credentials[publicKey]
	if !ok || credential.SiteID != siteID || !credential.Enabled {
		return Credential{}, ErrInvalidCredential
	}
	return credential, nil
}

type MemoryOnlineTracker struct {
	mu       sync.Mutex
	activity map[string]map[string]time.Time
	now      func() time.Time
}

func NewMemoryOnlineTracker() *MemoryOnlineTracker {
	return &MemoryOnlineTracker{
		activity: make(map[string]map[string]time.Time),
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (t *MemoryOnlineTracker) MarkActive(_ context.Context, siteID, visitorID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.activity[siteID] == nil {
		t.activity[siteID] = make(map[string]time.Time)
	}
	t.activity[siteID][visitorID] = t.now()
	return nil
}

type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, domain.EventEnvelope) (queue.PublishedMessage, error) {
	return queue.PublishedMessage{ID: "noop"}, nil
}
