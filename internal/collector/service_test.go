package collector

import (
	"context"
	"testing"

	"webanalytics/internal/domain"
	"webanalytics/internal/queue"
)

func TestCollectorAcceptsValidEventWithoutPostgres(t *testing.T) {
	publisher := &fakePublisher{}
	online := &fakeOnlineTracker{}
	service := NewService(ServiceDeps{
		Credentials: fakeCredentials{valid: true},
		Publisher:   publisher,
		Online:      online,
		Limiter:     AllowAllLimiter{},
	})

	result, err := service.Collect(context.Background(), CollectRequest{
		PublicKey: "pk_1",
		Type:      domain.EventTypePageView,
		SiteID:    "site_1",
		Visitor:   domain.Visitor{ID: "visitor_1", SessionID: "session_1"},
		Client: ClientMetadata{
			IP:        "203.0.113.1",
			UserAgent: "Mozilla/5.0 Test",
		},
	})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if !result.Accepted {
		t.Fatal("result not accepted")
	}
	if publisher.published.ID == "" {
		t.Fatal("event not published")
	}
	if publisher.published.Network.UserAgent != "Mozilla/5.0 Test" {
		t.Fatalf("user agent = %q", publisher.published.Network.UserAgent)
	}
	if online.visitorID != "visitor_1" {
		t.Fatalf("online visitor = %q, want visitor_1", online.visitorID)
	}
}

func TestCollectorRejectsInvalidCredentialBeforePublish(t *testing.T) {
	publisher := &fakePublisher{}
	service := NewService(ServiceDeps{
		Credentials: fakeCredentials{valid: false},
		Publisher:   publisher,
		Online:      &fakeOnlineTracker{},
		Limiter:     AllowAllLimiter{},
	})

	_, err := service.Collect(context.Background(), CollectRequest{
		PublicKey: "bad",
		Type:      domain.EventTypePageView,
		SiteID:    "site_1",
		Visitor:   domain.Visitor{ID: "visitor_1", SessionID: "session_1"},
	})
	if err == nil {
		t.Fatal("expected invalid credential error")
	}
	if publisher.published.ID != "" {
		t.Fatal("event was published for invalid credential")
	}
}

type fakeCredentials struct {
	valid bool
}

func (f fakeCredentials) Validate(context.Context, string, string) (Credential, error) {
	if !f.valid {
		return Credential{}, ErrInvalidCredential
	}
	return Credential{SiteID: "site_1", Enabled: true}, nil
}

type fakePublisher struct {
	published domain.EventEnvelope
}

func (f *fakePublisher) Publish(_ context.Context, event domain.EventEnvelope) (queue.PublishedMessage, error) {
	f.published = event
	return queue.PublishedMessage{ID: "msg_1"}, nil
}

type fakeOnlineTracker struct {
	visitorID string
}

func (f *fakeOnlineTracker) MarkActive(_ context.Context, siteID, visitorID string) error {
	f.visitorID = visitorID
	return nil
}
