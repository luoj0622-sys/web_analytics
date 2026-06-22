package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"webanalytics/internal/domain"
	"webanalytics/internal/store"
)

func TestEventStoreWritesRawEventsInOneBatch(t *testing.T) {
	exec := &fakeExecutor{}
	eventStore := NewEventStore(exec)

	err := eventStore.WriteRawBatch(context.Background(), []domain.EventEnvelope{
		{
			ID:         "evt_1",
			SiteID:     "00000000-0000-0000-0000-000000000001",
			Type:       domain.EventTypePageView,
			OccurredAt: time.Date(2026, 6, 22, 10, 30, 0, 0, time.UTC),
			Visitor:    domain.Visitor{ID: "visitor_1", SessionID: "session_1"},
		},
		{
			ID:         "evt_2",
			SiteID:     "00000000-0000-0000-0000-000000000001",
			Type:       domain.EventTypeHeartbeat,
			OccurredAt: time.Date(2026, 6, 22, 10, 31, 0, 0, time.UTC),
			Visitor:    domain.Visitor{ID: "visitor_1", SessionID: "session_1"},
		},
	})
	if err != nil {
		t.Fatalf("WriteRawBatch() error = %v", err)
	}
	if exec.calls != 1 {
		t.Fatalf("Exec calls = %d, want 1", exec.calls)
	}
	if !strings.Contains(exec.query, "INSERT INTO raw_events") {
		t.Fatalf("query missing raw insert: %s", exec.query)
	}
	if len(exec.args) == 0 {
		t.Fatal("expected query args")
	}
}

func TestStatsStoreUpsertsSiteStats(t *testing.T) {
	exec := &fakeExecutor{}
	statsStore := NewStatsStore(exec)

	err := statsStore.UpsertSiteStats(context.Background(), []store.SiteStat{
		{
			SiteID:         "00000000-0000-0000-0000-000000000001",
			Grain:          store.GrainHour,
			Bucket:         time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC),
			PageViews:      10,
			UniqueVisitors: 7,
			Sessions:       8,
			CustomEvents:   2,
		},
	})
	if err != nil {
		t.Fatalf("UpsertSiteStats() error = %v", err)
	}
	if !strings.Contains(exec.query, "INSERT INTO stats_site_hour") {
		t.Fatalf("query missing hour table: %s", exec.query)
	}
	if !strings.Contains(exec.query, "ON CONFLICT") {
		t.Fatalf("query missing upsert: %s", exec.query)
	}
}

type fakeExecutor struct {
	calls int
	query string
	args  []any
}

func (f *fakeExecutor) ExecContext(_ context.Context, query string, args ...any) error {
	f.calls++
	f.query = query
	f.args = args
	return nil
}
