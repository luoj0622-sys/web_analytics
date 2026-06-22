# Operations Runbook

## Scale Workers

1. Check `redis_stream_consumer_lag`, `redis_stream_pending_messages`, and `worker_batch_latency`.
2. Increase worker replicas while watching `postgres_write_latency`.
3. If PostgreSQL write latency rises sharply, reduce worker concurrency and inspect indexes, WAL volume, and partition health.

## Handle Redis Backlog

1. Confirm collectors are publishing successfully.
2. Inspect worker errors and failed batch counts.
3. Reclaim pending messages when a worker exits before acknowledgement.
4. Pause optional custom events if backlog continues to grow.

## Maintain PostgreSQL Partitions

1. Confirm future raw partitions exist for the next configured days.
2. Run partition creation before UTC day rollover if a partition is missing.
3. Verify raw retention actions detach, archive, or drop only partitions older than the configured retention window.
4. Keep raw-event indexes limited to site/time and BRIN time scans.

## RabbitMQ or Kafka Migration

1. Implement `EventPublisher` and `EventConsumer` for the new queue backend.
2. Run queue contract tests.
3. Deploy workers against the new consumer while collectors continue through the queue abstraction.
4. Cut over publishers after backlog and acknowledgement behavior are verified.

## ClickHouse Migration

1. Implement `EventStore` and `StatsStore` for ClickHouse.
2. Dual-write from workers or replay retained queue/archive data.
3. Keep PostgreSQL as the source of truth for users, sites, credentials, permissions, and settings.
4. Move dashboard analytical queries to ClickHouse after parity checks pass.
