package contracts

import (
	"context"
	"testing"

	"webanalytics/internal/domain"
	"webanalytics/internal/queue"
	"webanalytics/internal/store"
)

func TestFutureQueueBackendsCanSatisfyInterfaces(t *testing.T) {
	var _ queue.EventPublisher = futureRabbitMQPublisher{}
	var _ queue.EventConsumer = futureKafkaConsumer{}
}

func TestFutureClickHouseStoreCanSatisfyInterfaces(t *testing.T) {
	var _ store.EventStore = futureClickHouseEventStore{}
	var _ store.StatsStore = futureClickHouseStatsStore{}
}

type futureRabbitMQPublisher struct{}

func (futureRabbitMQPublisher) Publish(context.Context, domain.EventEnvelope) (queue.PublishedMessage, error) {
	return queue.PublishedMessage{}, nil
}

type futureKafkaConsumer struct{}

func (futureKafkaConsumer) ConsumeBatch(context.Context, queue.ConsumeOptions) (queue.Batch, error) {
	return queue.Batch{}, nil
}

func (futureKafkaConsumer) Ack(context.Context, []queue.MessageID) error {
	return nil
}

func (futureKafkaConsumer) Retry(context.Context, []queue.MessageID) error {
	return nil
}

func (futureKafkaConsumer) Lag(context.Context) (queue.Lag, error) {
	return queue.Lag{}, nil
}

type futureClickHouseEventStore struct{}

func (futureClickHouseEventStore) WriteRawBatch(context.Context, []domain.EventEnvelope) error {
	return nil
}

type futureClickHouseStatsStore struct{}

func (futureClickHouseStatsStore) UpsertSiteStats(context.Context, []store.SiteStat) error {
	return nil
}

func (futureClickHouseStatsStore) UpsertDimensionStats(context.Context, []store.DimensionStat) error {
	return nil
}

func (futureClickHouseStatsStore) QuerySiteStats(context.Context, store.SiteStatsQuery) (store.SiteStatsReport, error) {
	return store.SiteStatsReport{}, nil
}

func (futureClickHouseStatsStore) QueryDimensionStats(context.Context, store.DimensionStatsQuery) (store.DimensionStatsReport, error) {
	return store.DimensionStatsReport{}, nil
}
