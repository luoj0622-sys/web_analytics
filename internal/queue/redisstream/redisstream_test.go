package redisstream

import (
	"context"
	"testing"
	"time"

	"webanalytics/internal/domain"
	"webanalytics/internal/queue"
)

func TestPublisherWritesEventToRedisStream(t *testing.T) {
	client := &fakeCommander{}
	publisher := NewPublisher(client, PublisherConfig{Stream: "analytics:events"})

	msg, err := publisher.Publish(context.Background(), domain.EventEnvelope{
		ID:         "evt_1",
		SiteID:     "site_1",
		Type:       domain.EventTypePageView,
		OccurredAt: time.Date(2026, 6, 22, 10, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if msg.ID != queue.MessageID("174") {
		t.Fatalf("message id = %q, want 174", msg.ID)
	}
	if client.xaddStream != "analytics:events" {
		t.Fatalf("xadd stream = %q, want analytics:events", client.xaddStream)
	}
	if client.xaddValues["event_id"] != "evt_1" {
		t.Fatalf("event_id = %q, want evt_1", client.xaddValues["event_id"])
	}
}

func TestConsumerReadsAndAcknowledgesBatch(t *testing.T) {
	client := &fakeCommander{
		readMessages: []StreamMessage{{
			ID: "msg_1",
			Values: map[string]string{
				"event_id":    "evt_1",
				"site_id":     "site_1",
				"event_type":  "page_view",
				"occurred_at": "2026-06-22T10:30:00Z",
			},
		}},
	}
	consumer := NewConsumer(client, ConsumerConfig{
		Stream: "analytics:events",
		Group:  "analytics-workers",
		Name:   "worker-a",
	})

	batch, err := consumer.ConsumeBatch(context.Background(), queue.ConsumeOptions{
		MaxMessages: 10,
		BlockFor:    time.Millisecond,
	})
	if err != nil {
		t.Fatalf("ConsumeBatch() error = %v", err)
	}
	if len(batch.Messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(batch.Messages))
	}
	if batch.Messages[0].Event.ID != "evt_1" {
		t.Fatalf("event id = %q, want evt_1", batch.Messages[0].Event.ID)
	}

	if err := consumer.Ack(context.Background(), []queue.MessageID{"msg_1"}); err != nil {
		t.Fatalf("Ack() error = %v", err)
	}
	if client.acked[0] != "msg_1" {
		t.Fatalf("acked = %v, want msg_1", client.acked)
	}
}

type fakeCommander struct {
	xaddStream   string
	xaddValues   map[string]string
	readMessages []StreamMessage
	acked        []string
}

func (f *fakeCommander) XAdd(_ context.Context, stream string, values map[string]string) (string, error) {
	f.xaddStream = stream
	f.xaddValues = values
	return "174", nil
}

func (f *fakeCommander) XReadGroup(_ context.Context, _ ReadGroupArgs) ([]StreamMessage, error) {
	return f.readMessages, nil
}

func (f *fakeCommander) XAck(_ context.Context, _ string, _ string, ids []string) error {
	f.acked = append(f.acked, ids...)
	return nil
}

func (f *fakeCommander) XPending(_ context.Context, _ string, _ string) (PendingSummary, error) {
	return PendingSummary{}, nil
}
