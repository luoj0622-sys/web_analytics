# Observability

## Capacity Targets

- Stable target: 30 million events/day.
- Stretch target: 50 million events/day.
- Collector success path must remain independent of PostgreSQL availability.
- Dashboard historical freshness may degrade under queue backlog, but collectors should continue accepting valid events while Redis is healthy.

## Metrics

| Metric | Source | Purpose |
| --- | --- | --- |
| `collector_request_rate` | collector | Incoming event rate by site and event type |
| `collector_request_latency` | collector | Collector response latency |
| `queue_publish_latency` | collector | Time spent publishing to Redis Streams |
| `redis_stream_length` | queue | Total queued and retained stream messages |
| `redis_stream_consumer_lag` | queue | Worker lag by consumer group |
| `redis_stream_pending_messages` | queue | Messages read but not acknowledged |
| `worker_batch_size` | worker | Actual messages processed per batch |
| `worker_batch_latency` | worker | End-to-end worker processing time |
| `postgres_write_latency` | storage | Raw batch insert and aggregate upsert latency |
| `postgres_partition_create_failures` | retention | Failed future partition creation |
| `raw_partition_disk_growth` | storage | Daily raw partition storage growth |
| `failed_batches_total` | worker | Failed worker batches requiring retry |

## Alert Thresholds

- `redis_stream_consumer_lag` above 5 minutes for 10 minutes: scale workers or inspect PostgreSQL latency.
- `redis_stream_pending_messages` growing for 10 minutes: inspect failed batches and dead-letter behavior.
- `postgres_write_latency` P95 above 500 ms for raw batch writes: reduce batch concurrency or inspect indexes/WAL.
- Missing future partition within 24 hours: run partition creation job before midnight UTC.
- Raw partition storage growth above forecast: shorten raw retention or archive older partitions.

## Degradation Policy

1. Preserve collector acceptance for valid events.
2. Allow dashboard historical freshness to lag while queue backlog drains.
3. Temporarily pause non-critical custom-event processing if raw writes fall behind.
4. Scale workers horizontally before changing queue/storage implementations.
5. Consider RabbitMQ/Kafka or ClickHouse when sustained backlog cannot drain within the expected window.
