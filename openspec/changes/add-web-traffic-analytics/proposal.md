## Why

We need a self-hosted web traffic analytics system that can be embedded into owned websites, collect common traffic and visitor metrics, and serve real-time and historical dashboards without introducing heavyweight infrastructure at the start. The initial architecture should use Go, Redis, and PostgreSQL while stretching that stack toward 30 million events per day as a stable target and 50 million events per day as a stretch target.

## What Changes

- Add a JavaScript tracking integration for websites to report page views, heartbeats, sessions, referrers, UTM attributes, device/browser details, and common custom events.
- Add a Go-based collector that validates site credentials, records online activity, publishes events to a queue, and returns quickly without writing directly to PostgreSQL.
- Add Redis Streams as the initial event queue with producer/consumer semantics, consumer groups, acknowledgement, retry handling, and backlog observability.
- Add Go workers that consume events in batches, write raw events to PostgreSQL partitions, and update aggregate statistics through compressed batch upserts.
- Add PostgreSQL storage with daily raw-event partitions, aggregate tables, and hot/warm/cold retention policies.
- Add dashboard query capabilities backed by aggregate tables and Redis online state instead of scanning raw event data.
- Preserve migration exits for RabbitMQ/Kafka as future queue implementations and ClickHouse as a future analytics store.

## Capabilities

### New Capabilities

- `web-traffic-analytics`: Site tracking, event collection, real-time online users, traffic aggregation, reporting queries, partitioned PostgreSQL storage, retention, and future queue/analytics-store extension points.

### Modified Capabilities

- None.

## Impact

- Adds backend services for collection APIs, worker processing, site management APIs, and dashboard query APIs.
- Adds PostgreSQL schema objects for sites, tracking keys, raw event partitions, aggregate statistics, sessions, and retention metadata.
- Adds Redis usage for Streams-based event buffering, online user state, short-lived counters, and dashboard cache entries.
- Adds a browser JavaScript SDK for website integration.
- Introduces operational requirements for queue backlog monitoring, worker lag, PostgreSQL partition maintenance, retention jobs, and capacity testing.
