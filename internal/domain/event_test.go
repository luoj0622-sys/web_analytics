package domain

import (
	"testing"
	"time"
)

func TestEventEnvelopeAcceptsCommonAnalyticsFields(t *testing.T) {
	occurredAt := time.Date(2026, 6, 22, 10, 30, 0, 0, time.UTC)

	event := EventEnvelope{
		ID:         "evt_123",
		SiteID:     "site_123",
		Type:       EventTypePageView,
		OccurredAt: occurredAt,
		Visitor: Visitor{
			ID:        "visitor_123",
			SessionID: "session_123",
		},
		Page: Page{
			URL:      "https://example.com/pricing?utm_source=newsletter",
			Referrer: "https://search.example",
		},
		Campaign: Campaign{
			Source: "newsletter",
			Medium: "email",
		},
		Device: Device{
			Browser: "Safari",
			OS:      "macOS",
			Type:    "desktop",
		},
		Properties: map[string]any{"plan": "pro"},
	}

	if event.Type != EventTypePageView {
		t.Fatalf("Type = %q, want page_view", event.Type)
	}
	if event.Visitor.SessionID == "" {
		t.Fatal("SessionID is empty")
	}
	if event.OccurredAt != occurredAt {
		t.Fatalf("OccurredAt = %s, want %s", event.OccurredAt, occurredAt)
	}
}
