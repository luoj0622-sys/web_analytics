# Web Analytics

Self-hosted web traffic analytics for owned websites. The first implementation targets a simple operational stack:

- Go services for collection, workers, and dashboard APIs
- Redis Streams for queue buffering and worker consumption
- Redis online state for real-time active visitors
- PostgreSQL for business data, raw-event partitions, and aggregate reporting tables
- JavaScript SDK for pageview, heartbeat, and custom-event collection

The design target is 30 million events/day as the stable capacity goal and 50 million events/day as the stretch goal, assuming queue buffering, batch writes, partitioning, aggregate-first dashboard queries, and bounded raw retention.

## Current Status

This repository contains a tested first implementation skeleton. The code defines the service boundaries, migrations, SDK, collector, worker, dashboard, retention planner, and operational documents. Some production integrations are intentionally still behind interfaces:

- Collector entrypoint currently uses in-memory credentials and a no-op publisher.
- Dashboard entrypoint currently uses in-memory/empty readers.
- PostgreSQL and Redis adapters exist as boundaries, but service wiring to real clients still needs to be completed.
- RabbitMQ/Kafka and ClickHouse are preserved as future exits through queue and storage interfaces.

## Configuration

Runtime information is centralized in [config/local.example.yaml](/Users/a1-6/Documents/workspace/Web_Analytics/config/local.example.yaml).

Key sections:

- `postgres`: PostgreSQL DSN, connection pool, migration directory
- `redis`: Redis address, DB, online/cache key prefixes
- `queue`: Redis Streams driver, stream name, consumer group, retry/dead-letter settings
- `worker`: batch size and flush/idle behavior
- `retention`: raw and aggregate retention windows
- `capacity`: 30 million stable and 50 million stretch daily event targets
- `future_backends`: RabbitMQ/Kafka and ClickHouse migration exits

The current Go config loader reads environment variables from [.env.example](/Users/a1-6/Documents/workspace/Web_Analytics/.env.example). The YAML file is the canonical local runtime config reference and should be mirrored into environment variables until file-based config loading is added.

## Local Dependencies

Start PostgreSQL and Redis:

```bash
docker compose up -d postgres redis
```

Default local services:

- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Collector: `:8080`
- Dashboard: `:8081`

## Service Commands

```bash
go run ./cmd/collector
go run ./cmd/dashboard
go run ./cmd/worker
```

Health checks:

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8081/healthz
```

## SDK

The SDK core lives in [sdk/tracker.mjs](/Users/a1-6/Documents/workspace/Web_Analytics/sdk/tracker.mjs).

It supports:

- visitor/session bootstrap
- pageview payloads
- heartbeat payloads
- custom events with structured properties

## Tests

Go tests:

```bash
GOCACHE=/Users/a1-6/Documents/workspace/Web_Analytics/.cache/go-build go test ./...
```

SDK tests:

```bash
npm run test:sdk
```

## Capacity And Operations

Operational guidance lives in:

- [docs/observability.md](/Users/a1-6/Documents/workspace/Web_Analytics/docs/observability.md)
- [docs/runbook.md](/Users/a1-6/Documents/workspace/Web_Analytics/docs/runbook.md)
- [tools/loadtest.mjs](/Users/a1-6/Documents/workspace/Web_Analytics/tools/loadtest.mjs)

The intended high-volume path is:

```text
JS SDK -> Go Collector -> Redis Streams -> Go Workers -> PostgreSQL raw partitions + aggregate tables
```

Dashboard APIs should read Redis online state and aggregate tables, not raw-event partitions.

## Future Backend Exits

The implementation keeps these exits explicit:

- Queue: Redis Streams now, RabbitMQ or Kafka later through `EventPublisher` and `EventConsumer`.
- Analytics store: PostgreSQL now, ClickHouse later through `EventStore` and `StatsStore`.

PostgreSQL should remain the system of record for users, sites, credentials, permissions, and configuration even if ClickHouse becomes the analytical event store.
