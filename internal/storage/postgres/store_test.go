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
			IPCount:        6,
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
	if !strings.Contains(exec.query, "ip_count") {
		t.Fatalf("query missing ip_count: %s", exec.query)
	}
}

func TestStatsStoreQueriesSiteStats(t *testing.T) {
	exec := &fakeExecutor{
		rows: &fakeRows{values: [][]any{{
			"00000000-0000-0000-0000-000000000001",
			time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC),
			int64(10),
			int64(6),
			int64(7),
			int64(8),
			int64(2),
		}}},
	}
	statsStore := NewStatsStore(exec)

	report, err := statsStore.QuerySiteStats(context.Background(), store.SiteStatsQuery{
		SiteID: "00000000-0000-0000-0000-000000000001",
		Grain:  store.GrainHour,
		From:   time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("QuerySiteStats() error = %v", err)
	}
	if !strings.Contains(exec.query, "FROM stats_site_hour") {
		t.Fatalf("query missing hour table: %s", exec.query)
	}
	if len(report.Rows) != 1 || report.Rows[0].IPCount != 6 {
		t.Fatalf("rows = %+v, want ip_count 6", report.Rows)
	}
}

func TestStatsStoreQueriesDimensionStats(t *testing.T) {
	exec := &fakeExecutor{
		rows: &fakeRows{values: [][]any{{
			"00000000-0000-0000-0000-000000000001",
			time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC),
			"https://example.com/docs",
			int64(10),
			int64(6),
			int64(7),
			int64(8),
			int64(0),
		}}},
	}
	statsStore := NewStatsStore(exec)

	report, err := statsStore.QueryDimensionStats(context.Background(), store.DimensionStatsQuery{
		SiteID:    "00000000-0000-0000-0000-000000000001",
		Dimension: store.DimensionReferrer,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("QueryDimensionStats() error = %v", err)
	}
	if !strings.Contains(exec.query, "FROM stats_referrer_day") {
		t.Fatalf("query missing referrer table: %s", exec.query)
	}
	if len(report.Rows) != 1 || report.Rows[0].Key != "https://example.com/docs" || report.Rows[0].IPCount != 6 {
		t.Fatalf("rows = %+v, want referrer row with ip_count 6", report.Rows)
	}
}

type fakeExecutor struct {
	calls int
	query string
	args  []any
	rows  Rows
}

func (f *fakeExecutor) ExecContext(_ context.Context, query string, args ...any) error {
	f.calls++
	f.query = query
	f.args = args
	return nil
}

func (f *fakeExecutor) QueryContext(_ context.Context, query string, args ...any) (Rows, error) {
	f.calls++
	f.query = query
	f.args = args
	return f.rows, nil
}

type fakeRows struct {
	values [][]any
	index  int
}

func (r *fakeRows) Next() bool {
	return r.index < len(r.values)
}

func (r *fakeRows) Scan(dest ...any) error {
	row := r.values[r.index]
	r.index++
	for i := range dest {
		switch target := dest[i].(type) {
		case *string:
			*target = row[i].(string)
		case *time.Time:
			*target = row[i].(time.Time)
		case *int64:
			*target = row[i].(int64)
		}
	}
	return nil
}

func (r *fakeRows) Close() error {
	return nil
}

func (r *fakeRows) Err() error {
	return nil
}
