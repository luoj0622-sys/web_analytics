package queue

import (
	"context"
	"time"

	"webanalytics/internal/domain"
)

type MessageID string

type PublishedMessage struct {
	ID MessageID
}

type Message struct {
	ID    MessageID
	Event domain.EventEnvelope
}

type Batch struct {
	Messages []Message
}

type ConsumeOptions struct {
	MaxMessages int
	BlockFor    time.Duration
}

type Lag struct {
	StreamLength    int64
	PendingMessages int64
	ConsumerLag     int64
	FailedMessages  int64
}

type EventPublisher interface {
	Publish(context.Context, domain.EventEnvelope) (PublishedMessage, error)
}

type EventConsumer interface {
	ConsumeBatch(context.Context, ConsumeOptions) (Batch, error)
	Ack(context.Context, []MessageID) error
	Retry(context.Context, []MessageID) error
	Lag(context.Context) (Lag, error)
}
