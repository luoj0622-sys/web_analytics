package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"webanalytics/internal/domain"
	"webanalytics/internal/queue"
	"webanalytics/internal/store"
)

func TestProcessorWritesRawAggregatesAndAcknowledges(t *testing.T) {
	consumer := &fakeConsumer{
		batch: queue.Batch{Messages: []queue.Message{
			{ID: "msg_1", Event: event("evt_1", domain.EventTypePageView, "/pricing")},
			{ID: "msg_2", Event: event("evt_2", domain.EventTypeCustom, "/pricing")},
		}},
	}
	events := &fakeEventStore{}
	stats := &fakeStatsStore{}
	processor := NewProcessor(ProcessorDeps{
		Consumer: consumer,
		Events:   events,
		Stats:    stats,
		Config:   Config{BatchSize: 1000, IdleTimeout: time.Millisecond},
	})

	if err := processor.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}

	if len(events.written) != 2 {
		t.Fatalf("raw written = %d, want 2", len(events.written))
	}
	if len(stats.upserts) == 0 {
		t.Fatal("expected aggregate upserts")
	}
	if stats.upserts[0].IPCount != 1 {
		t.Fatalf("site ip count = %d, want 1", stats.upserts[0].IPCount)
	}
	if len(stats.dimensionUpserts) == 0 {
		t.Fatal("expected dimension aggregate upserts")
	}
	if stats.dimensionUpserts[0].IPCount != 1 {
		t.Fatalf("dimension ip count = %d, want 1", stats.dimensionUpserts[0].IPCount)
	}
	if len(consumer.acked) != 2 {
		t.Fatalf("acked = %v, want 2 ids", consumer.acked)
	}
}

func TestProcessorRetriesBatchWhenRawWriteFails(t *testing.T) {
	consumer := &fakeConsumer{
		batch: queue.Batch{Messages: []queue.Message{{ID: "msg_1", Event: event("evt_1", domain.EventTypePageView, "/pricing")}}},
	}
	events := &fakeEventStore{err: errors.New("db down")}
	processor := NewProcessor(ProcessorDeps{
		Consumer: consumer,
		Events:   events,
		Stats:    &fakeStatsStore{},
		Config:   Config{BatchSize: 1000, IdleTimeout: time.Millisecond},
	})

	if err := processor.RunOnce(context.Background()); err == nil {
		t.Fatal("expected raw write error")
	}
	if len(consumer.retried) != 1 {
		t.Fatalf("retried = %v, want msg_1", consumer.retried)
	}
	if len(consumer.acked) != 0 {
		t.Fatalf("acked = %v, want none", consumer.acked)
	}
}

func event(id string, typ domain.EventType, path string) domain.EventEnvelope {
	return domain.EventEnvelope{
		ID:         id,
		SiteID:     "site_1",
		Type:       typ,
		Name:       "signup",
		OccurredAt: time.Date(2026, 6, 22, 10, 35, 0, 0, time.UTC),
		Visitor:    domain.Visitor{ID: "visitor_1", SessionID: "session_1"},
		Page:       domain.Page{Path: path},
		Network:    domain.Network{IP: "ip_hash_1"},
	}
}

type fakeConsumer struct {
	batch   queue.Batch
	acked   []queue.MessageID
	retried []queue.MessageID
}

func (f *fakeConsumer) ConsumeBatch(context.Context, queue.ConsumeOptions) (queue.Batch, error) {
	return f.batch, nil
}

func (f *fakeConsumer) Ack(_ context.Context, ids []queue.MessageID) error {
	f.acked = append(f.acked, ids...)
	return nil
}

func (f *fakeConsumer) Retry(_ context.Context, ids []queue.MessageID) error {
	f.retried = append(f.retried, ids...)
	return nil
}

func (f *fakeConsumer) Lag(context.Context) (queue.Lag, error) {
	return queue.Lag{}, nil
}

type fakeEventStore struct {
	written []domain.EventEnvelope
	err     error
}

func (f *fakeEventStore) WriteRawBatch(_ context.Context, events []domain.EventEnvelope) error {
	if f.err != nil {
		return f.err
	}
	f.written = append(f.written, events...)
	return nil
}

type fakeStatsStore struct {
	upserts          []store.SiteStat
	dimensionUpserts []store.DimensionStat
}

func (f *fakeStatsStore) UpsertSiteStats(_ context.Context, stats []store.SiteStat) error {
	f.upserts = append(f.upserts, stats...)
	return nil
}

func (f *fakeStatsStore) UpsertDimensionStats(_ context.Context, stats []store.DimensionStat) error {
	f.dimensionUpserts = append(f.dimensionUpserts, stats...)
	return nil
}

func (f *fakeStatsStore) QuerySiteStats(context.Context, store.SiteStatsQuery) (store.SiteStatsReport, error) {
	return store.SiteStatsReport{}, nil
}

func (f *fakeStatsStore) QueryDimensionStats(context.Context, store.DimensionStatsQuery) (store.DimensionStatsReport, error) {
	return store.DimensionStatsReport{}, nil
}
