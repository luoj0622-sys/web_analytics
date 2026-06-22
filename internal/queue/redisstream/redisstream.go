package redisstream

import (
	"context"
	"time"

	"webanalytics/internal/domain"
	"webanalytics/internal/queue"
)

type Commander interface {
	XAdd(ctx context.Context, stream string, values map[string]string) (string, error)
	XReadGroup(ctx context.Context, args ReadGroupArgs) ([]StreamMessage, error)
	XAck(ctx context.Context, stream, group string, ids []string) error
	XPending(ctx context.Context, stream, group string) (PendingSummary, error)
}

type PublisherConfig struct {
	Stream string
}

type ConsumerConfig struct {
	Stream string
	Group  string
	Name   string
}

type ReadGroupArgs struct {
	Stream   string
	Group    string
	Consumer string
	Count    int
	BlockFor time.Duration
}

type StreamMessage struct {
	ID     string
	Values map[string]string
}

type PendingSummary struct {
	Count int64
}

type Publisher struct {
	client Commander
	cfg    PublisherConfig
}

func NewPublisher(client Commander, cfg PublisherConfig) *Publisher {
	return &Publisher{client: client, cfg: cfg}
}

func (p *Publisher) Publish(ctx context.Context, event domain.EventEnvelope) (queue.PublishedMessage, error) {
	id, err := p.client.XAdd(ctx, p.cfg.Stream, encodeEvent(event))
	if err != nil {
		return queue.PublishedMessage{}, err
	}
	return queue.PublishedMessage{ID: queue.MessageID(id)}, nil
}

type Consumer struct {
	client Commander
	cfg    ConsumerConfig
}

func NewConsumer(client Commander, cfg ConsumerConfig) *Consumer {
	return &Consumer{client: client, cfg: cfg}
}

func (c *Consumer) ConsumeBatch(ctx context.Context, opts queue.ConsumeOptions) (queue.Batch, error) {
	messages, err := c.client.XReadGroup(ctx, ReadGroupArgs{
		Stream:   c.cfg.Stream,
		Group:    c.cfg.Group,
		Consumer: c.cfg.Name,
		Count:    opts.MaxMessages,
		BlockFor: opts.BlockFor,
	})
	if err != nil {
		return queue.Batch{}, err
	}

	batch := queue.Batch{Messages: make([]queue.Message, 0, len(messages))}
	for _, message := range messages {
		event, err := decodeEvent(message.Values)
		if err != nil {
			return queue.Batch{}, err
		}
		batch.Messages = append(batch.Messages, queue.Message{
			ID:    queue.MessageID(message.ID),
			Event: event,
		})
	}
	return batch, nil
}

func (c *Consumer) Ack(ctx context.Context, ids []queue.MessageID) error {
	return c.client.XAck(ctx, c.cfg.Stream, c.cfg.Group, messageIDs(ids))
}

func (c *Consumer) Retry(context.Context, []queue.MessageID) error {
	return nil
}

func (c *Consumer) Lag(ctx context.Context) (queue.Lag, error) {
	pending, err := c.client.XPending(ctx, c.cfg.Stream, c.cfg.Group)
	if err != nil {
		return queue.Lag{}, err
	}
	return queue.Lag{PendingMessages: pending.Count}, nil
}

func encodeEvent(event domain.EventEnvelope) map[string]string {
	values := map[string]string{
		"event_id":    event.ID,
		"site_id":     event.SiteID,
		"event_type":  string(event.Type),
		"event_name":  event.Name,
		"occurred_at": event.OccurredAt.UTC().Format(time.RFC3339Nano),
		"visitor_id":  event.Visitor.ID,
		"session_id":  event.Visitor.SessionID,
		"page_url":    event.Page.URL,
		"page_path":   event.Page.Path,
		"referrer":    event.Page.Referrer,
		"utm_source":  event.Campaign.Source,
		"utm_medium":  event.Campaign.Medium,
		"browser":     event.Device.Browser,
		"os":          event.Device.OS,
		"device_type": event.Device.Type,
	}
	if !event.ReceivedAt.IsZero() {
		values["received_at"] = event.ReceivedAt.UTC().Format(time.RFC3339Nano)
	}
	return values
}

func decodeEvent(values map[string]string) (domain.EventEnvelope, error) {
	occurredAt, err := parseTime(values["occurred_at"])
	if err != nil {
		return domain.EventEnvelope{}, err
	}
	receivedAt, err := parseOptionalTime(values["received_at"])
	if err != nil {
		return domain.EventEnvelope{}, err
	}
	return domain.EventEnvelope{
		ID:         values["event_id"],
		SiteID:     values["site_id"],
		Type:       domain.EventType(values["event_type"]),
		Name:       values["event_name"],
		OccurredAt: occurredAt,
		ReceivedAt: receivedAt,
		Visitor: domain.Visitor{
			ID:        values["visitor_id"],
			SessionID: values["session_id"],
		},
		Page: domain.Page{
			URL:      values["page_url"],
			Path:     values["page_path"],
			Referrer: values["referrer"],
		},
		Campaign: domain.Campaign{
			Source: values["utm_source"],
			Medium: values["utm_medium"],
		},
		Device: domain.Device{
			Browser: values["browser"],
			OS:      values["os"],
			Type:    values["device_type"],
		},
	}, nil
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}

func parseOptionalTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return parseTime(value)
}

func messageIDs(ids []queue.MessageID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	return out
}
