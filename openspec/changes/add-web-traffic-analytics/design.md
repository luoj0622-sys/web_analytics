## Context

The system will provide self-hosted traffic analytics for owned websites. The first production architecture must stay operationally simple by using Go, Redis, and PostgreSQL, while still being able to handle high event volume: 30 million events per day as a stable target and 50 million events per day as a stretch target.

The core load pattern is write-heavy. Website pages send analytics events to a collector; the collector must protect user-facing sites by returning quickly and avoiding direct PostgreSQL writes. PostgreSQL must be used carefully as an initial analytics store through daily partitioning, batch ingestion, aggregate tables, and retention policies. Redis must provide online-user state, short-lived hot counters, and the initial queue implementation.

## Goals / Non-Goals

**Goals:**

- Provide website tracking through a JavaScript SDK and collector API.
- Support common analytics metrics: page views, unique visitors, sessions, referrers, UTM campaigns, popular pages, devices, browsers, operating systems, custom events, and real-time online users.
- Use Redis Streams as the initial queue with producer/consumer semantics, acknowledgements, retries, and lag monitoring.
- Use PostgreSQL as the initial raw-event and aggregate store with daily partitioning and hot/warm/cold retention.
- Keep dashboard queries fast by reading aggregate tables and Redis online state instead of scanning raw events.
- Design queue and analytics-store boundaries so RabbitMQ/Kafka and ClickHouse can be introduced later without rewriting collector and worker business logic.

**Non-Goals:**

- Do not introduce ClickHouse, Kafka, or RabbitMQ in the first implementation.
- Do not support unlimited raw-event retention in PostgreSQL.
- Do not support full user-journey replay, heatmaps, session recordings, or A/B testing in the first implementation.
- Do not guarantee exact real-time aggregation under backlog; bounded eventual consistency is acceptable for historical metrics.

## Decisions

### Use Redis Streams as the initial queue

Collector instances will publish normalized event envelopes to Redis Streams. Worker instances will consume through Redis consumer groups, acknowledge successful batches, retry pending messages, and emit lag metrics.

Alternatives considered:

- Direct PostgreSQL writes from collectors: simpler, but database latency and outages would directly affect tracking requests and high-traffic spikes.
- Redis List: simpler than Streams, but weaker acknowledgement, retry, lag, and multi-consumer semantics.
- RabbitMQ/Kafka now: more scalable or durable for some cases, but adds operational complexity before the first version needs it.

### Keep queue access behind publisher and consumer interfaces

The collector will depend on an `EventPublisher` abstraction. Workers will depend on an `EventConsumer` abstraction. The first implementation will use Redis Streams, but the interface must be compatible with RabbitMQ/Kafka semantics: publish, consume batch, ack, retry/dead-letter, and expose lag/backlog metrics.

This preserves a future migration path:

```text
RedisStreamPublisher -> RabbitMQPublisher or KafkaPublisher
RedisStreamConsumer  -> RabbitMQConsumer  or KafkaConsumer
```

### Use PostgreSQL daily partitions for raw events

Raw analytics events will be stored in a partitioned PostgreSQL table, partitioned by event date. Daily partitions keep each partition manageable at 30-50 million events per day, allow cheap retention through detach/drop/archive operations, and support targeted maintenance.

The raw event table must have a narrow write path. It should avoid broad indexing. Required indexes should be limited to time/site access patterns such as `(site_id, event_time)` and BRIN indexes for time-ordered scans.

### Query aggregate tables, not raw events

Workers will aggregate events in memory per batch and write compressed updates to statistics tables. Dashboard APIs will read from Redis for real-time online state and from aggregate tables for historical reporting.

Representative aggregate tables:

- `stats_site_minute`, `stats_site_hour`, `stats_site_day`
- `stats_page_hour`, `stats_page_day`
- `stats_referrer_day`
- `stats_device_day`
- `stats_geo_day`
- `stats_event_day`

Raw events are for short-term detail, troubleshooting, and recomputation. They are not the primary source for routine dashboard queries.

### Use hot/warm/cold retention policies

The first implementation will include lifecycle jobs for partition creation, retention, and archival.

Recommended defaults:

- Redis hot data: minutes to hours.
- PostgreSQL hot raw data: recent 7 days.
- PostgreSQL warm raw data: 30-60 days total raw retention.
- Minute aggregates: 7-15 days.
- Hour aggregates: 180-365 days.
- Day aggregates: long-term retention.

Cold raw archival may export detached partitions or compressed event batches to an external storage location later. The first version should isolate the retention workflow so storage backends can be added.

### Preserve ClickHouse as a future analytics-store exit

PostgreSQL is the first analytics store, but analytics writes and reads must be isolated behind store interfaces:

```text
EventStore
  PostgresEventStore
  ClickHouseEventStore later

StatsStore
  PostgresStatsStore
  ClickHouseStatsStore later
```

PostgreSQL should continue to own business data such as users, sites, credentials, teams, and settings even if ClickHouse later owns high-volume raw events and analytical queries.

### Treat 30 million/day as stable and 50 million/day as stretch

The design target is 30 million events per day under normal operation. The stretch target is 50 million events per day with explicit constraints:

- Collectors must not synchronously write to PostgreSQL.
- Workers must batch raw inserts and aggregate upserts.
- Raw-event retention must stay bounded.
- Dashboard queries must use aggregates.
- Backlog, worker lag, PostgreSQL write latency, WAL volume, disk growth, and partition maintenance must be observable.

## Risks / Trade-offs

- PostgreSQL write amplification could exceed capacity -> Batch insert raw events, aggregate in memory before upsert, keep raw indexes minimal, and use daily partitions.
- Redis queue backlog could grow during PostgreSQL slowdowns -> Monitor stream length and consumer lag, scale workers horizontally, apply backpressure alerts, and define retention/dead-letter behavior.
- Raw-event storage can grow too quickly at 50 million events per day -> Default raw retention to 30 days, allow 60 days only with storage validation, and retain long-term aggregates instead of raw detail.
- Dashboard results may lag during traffic spikes -> Accept bounded eventual consistency for historical aggregates while keeping online users in Redis real-time.
- Future migration to ClickHouse or RabbitMQ could be expensive if implementation details leak -> Enforce queue and store abstractions from the first implementation.
- Partition maintenance failures could degrade writes or retention -> Add scheduled partition creation, retention jobs, health checks, and operational alerts.

## Migration Plan

1. Implement the Redis Streams + PostgreSQL path behind queue and store interfaces.
2. Deploy with low event volume and validate collector latency, queue lag, worker throughput, and aggregate correctness.
3. Enable daily partition creation and retention automation before production traffic.
4. Run load tests at expected peak rates for 30 million/day and stretch rates for 50 million/day.
5. If PostgreSQL becomes the bottleneck, introduce ClickHouse by dual-writing raw events from workers or replaying retained queue/archive data.
6. If Redis Streams becomes the bottleneck or durability requirement increases, introduce RabbitMQ/Kafka by implementing the existing queue interfaces.

Rollback strategy: disable tracking script loading or reject non-essential event types, pause custom-event ingestion, scale down worker consumers safely, and continue serving dashboards from the last successful aggregate state.

## Open Questions

- What exact raw-event retention default should ship first: 30 days or 60 days?
- Should cold raw archives be part of the first implementation or a follow-up capability?
- Which custom events should be considered first-class in the initial dashboard?
- What deployment topology will be used for PostgreSQL and Redis in production: single node, managed service, or replicated cluster?
