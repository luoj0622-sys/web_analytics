package queue

import (
	"context"
	"testing"
	"time"

	"webanalytics/internal/domain"
)

func TestQueueInterfacesSupportFutureBackends(t *testing.T) {
	var publisher EventPublisher = noopPublisher{}
	var consumer EventConsumer = noopConsumer{}

	ctx := context.Background()
	msg, err := publisher.Publish(ctx, domain.EventEnvelope{ID: "evt_1"})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if msg.ID == "" {
		t.Fatal("published message id is empty")
	}

	batch, err := consumer.ConsumeBatch(ctx, ConsumeOptions{MaxMessages: 10, BlockFor: time.Millisecond})
	if err != nil {
		t.Fatalf("ConsumeBatch() error = %v", err)
	}
	if len(batch.Messages) != 0 {
		t.Fatalf("len(batch.Messages) = %d, want 0", len(batch.Messages))
	}
	if err := consumer.Ack(ctx, []MessageID{"msg_1"}); err != nil {
		t.Fatalf("Ack() error = %v", err)
	}
}

type noopPublisher struct{}

func (noopPublisher) Publish(context.Context, domain.EventEnvelope) (PublishedMessage, error) {
	return PublishedMessage{ID: "msg_1"}, nil
}

type noopConsumer struct{}

func (noopConsumer) ConsumeBatch(context.Context, ConsumeOptions) (Batch, error) {
	return Batch{}, nil
}

func (noopConsumer) Ack(context.Context, []MessageID) error {
	return nil
}

func (noopConsumer) Retry(context.Context, []MessageID) error {
	return nil
}

func (noopConsumer) Lag(context.Context) (Lag, error) {
	return Lag{}, nil
}
