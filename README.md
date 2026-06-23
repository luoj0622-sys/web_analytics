# Web Analytics

## 中文说明

Web Analytics 是一个面向自有网站的自托管流量分析系统。项目目标是提供从前端埋点、事件采集、队列缓冲、后台消费、数据存储到统计查询的一套基础实现。

当前实现采用的核心组件：

- Go 服务：负责事件采集、后台消费任务和 Dashboard API
- Redis Streams：作为事件队列，缓冲采集到的访问事件
- Redis 在线状态：用于记录实时活跃访客
- PostgreSQL：存储业务数据、原始事件分区表和聚合统计表
- JavaScript SDK：支持页面访问、心跳和自定义事件上报

设计容量目标：

- 稳定目标：每天 3000 万事件
- 扩展目标：每天 5000 万事件

该目标依赖队列缓冲、批量写入、分区表、优先查询聚合表以及受控的原始数据保留策略。

### 当前状态

本仓库目前是一个经过测试的第一版实现骨架，已经包含服务边界、数据库迁移、SDK、采集服务、Worker、Dashboard、保留策略规划器和运维文档。部分生产集成仍然通过接口保留：

- Collector 入口当前使用内存凭证和空发布器
- Dashboard 入口当前使用内存或空读取器
- PostgreSQL 和 Redis 适配器已经定义边界，但真实客户端接入还需要继续完善
- RabbitMQ、Kafka 和 ClickHouse 作为后续替换方案，通过队列和存储接口保留扩展出口

### 本地依赖

启动 PostgreSQL 和 Redis：

```bash
docker compose up -d postgres redis
```

默认本地服务：

- PostgreSQL：`localhost:5432`
- Redis：`localhost:6379`
- Collector：`:8080`
- Dashboard：`:8081`

### 服务启动

```bash
go run ./cmd/collector
go run ./cmd/dashboard
go run ./cmd/worker
```

健康检查：

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8081/healthz
```

### 测试

Go 测试：

```bash
GOCACHE=/Users/a1-6/Documents/workspace/Web_Analytics/.cache/go-build go test ./...
```

SDK 测试：

```bash
npm run test:sdk
```

### 数据流

```text
JS SDK -> Go Collector -> Redis Streams -> Go Workers -> PostgreSQL 原始事件分区表 + 聚合统计表
```

Dashboard API 应优先读取 Redis 在线状态和聚合统计表，而不是直接查询原始事件分区表。

### 后续扩展方向

- 队列：当前使用 Redis Streams，后续可通过 `EventPublisher` 和 `EventConsumer` 切换到 RabbitMQ 或 Kafka
- 分析存储：当前使用 PostgreSQL，后续可通过 `EventStore` 和 `StatsStore` 接入 ClickHouse

即使未来引入 ClickHouse 作为分析事件存储，PostgreSQL 仍建议保留为用户、站点、凭证、权限和配置的系统记录库。

## English

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
