# Web Analytics

自托管的网站流量分析系统，从前端埋点、事件采集、队列缓冲、后台消费、数据存储到看板查询提供一套完整的基础实现。

Self-hosted web traffic analytics for your own websites: JavaScript SDK → collector → queue → workers → storage → dashboard.

---

## 中文说明

### 系统组成

| 组件 | 职责 |
| --- | --- |
| **JavaScript SDK** (`sdk/tracker.mjs`) | 页面访问、心跳、自定义事件上报 |
| **Collector** (`cmd/collector`) | 接收事件、校验凭证、限流、写入队列、标记在线访客 |
| **Worker** (`cmd/worker`) | 从队列批量消费事件，写入原始分区表与聚合统计表 |
| **Dashboard** (`cmd/dashboard`) | 提供 JSON 统计 API，并托管浏览器看板页面 |
| **Redis Streams** | 事件队列，缓冲采集到的访问事件 |
| **Redis 在线状态** | 记录实时活跃访客 |
| **PostgreSQL** | 业务数据、原始事件分区表、聚合统计表 |

设计容量目标：稳定 3000 万事件/天，扩展 5000 万事件/天，依赖队列缓冲、批量写入、分区表、聚合优先查询与受控的原始数据保留策略。

### 数据流

```text
JS SDK ──POST /collect──▶ Collector ──▶ Redis Streams ──▶ Workers ──▶ PostgreSQL
                              │                                         (原始分区表 + 聚合表)
                              └──▶ Redis 在线状态
                                                  Dashboard ──读取──▶ Redis 在线 + 聚合表
```

看板 API 优先读取 Redis 在线状态和聚合统计表，而不是直接查询原始事件分区表。

### 功能

- 核心指标：浏览量 PV、IP 数、访问次数 Session、活跃访客 UV、活跃访客跨天加和、累计访客 UV、融合访客 UV
- 指标趋势折线图（按分钟 / 小时 / 天聚合）
- 来路域名、来路页面维度报表（PV / IP / UV / Session）
- 多维度查询：页面、来路、UTM、设备、地理、自定义事件
- 实时在线访客统计
- 站点创建 API 与多站点隔离（按 `site_id` + `public_key`）

---

### 端口

启动方式不同，默认端口不同。

| 服务 | `go run` 本地默认 | docker compose | 环境变量 |
| --- | --- | --- | --- |
| Collector | `:8080` | `:8085` | `WA_COLLECTOR_ADDR` |
| Dashboard | `:8081` | `:8086` | `WA_DASHBOARD_ADDR` |
| PostgreSQL | `localhost:5432` | `localhost:5432` | `WA_POSTGRES_DSN` |
| Redis | `localhost:6379` | `localhost:6379` | `WA_REDIS_ADDR` |

> docker-compose.yml 通过环境变量把端口改成了 8085 / 8086。如果你用 `go run` 直接启动，端口是 8080 / 8081。下文示例以本地 `go run`（8080 / 8081）为准。

---

### 怎么用

#### 1. 启动依赖

```bash
docker compose up -d postgres redis
```

执行数据库迁移（按文件名顺序应用 `migrations/*.sql`）：

```bash
for f in migrations/*.sql; do
  psql "postgres://webanalytics:webanalytics@localhost:5432/webanalytics?sslmode=disable" -f "$f"
done
```

#### 2. 启动服务

本地分别启动三个服务：

```bash
go run ./cmd/collector   # 监听 :8080
go run ./cmd/dashboard   # 监听 :8081
go run ./cmd/worker
```

或用 docker compose 一键启动全部（端口为 8085 / 8086）：

```bash
docker compose up -d
```

健康检查：

```bash
curl -s http://localhost:8080/healthz   # collector
curl -s http://localhost:8081/healthz   # dashboard
```

#### 3. 打开看板

浏览器访问 dashboard 地址：

- 本地 `go run`：`http://localhost:8081/`
- docker compose：`http://localhost:8086/`

页面顶部输入站点 ID、选择粒度和时间范围，点击「刷新」即可查看指标、趋势图和来路报表。

> 配置说明：当前 Collector / Dashboard 入口使用内存凭证和空读取器作为骨架实现，真实的 PostgreSQL / Redis 客户端接入仍在完善中。看板页面调用真实 API，但返回数据依赖后端读取器的接入程度。

---

### 怎么埋点

#### 引入 SDK

SDK 是一个零依赖的 ES Module，位于 `sdk/tracker.mjs`，把它部署到你的静态资源目录后引入：

```html
<script type="module">
  import { createTracker } from "/sdk/tracker.mjs";

  const tracker = createTracker({
    siteId: "site_1",                          // 站点 ID
    publicKey: "pk_xxx",                        // 站点公钥（采集凭证）
    collectUrl: "http://localhost:8080/collect" // Collector 采集地址
  });

  // 页面访问
  tracker.trackPageView();

  // 心跳（用于在线访客统计，建议定时调用）
  setInterval(() => tracker.heartbeat(), 30_000);

  // 自定义事件
  tracker.track("signup", { plan: "pro", price: 99 });
</script>
```

#### SDK 接口

| 方法 | 上报事件类型 | 说明 |
| --- | --- | --- |
| `trackPageView()` | `page_view` | 上报当前页面访问 |
| `heartbeat()` | `heartbeat` | 上报心跳，维持在线访客状态 |
| `track(name, properties)` | `custom` | 上报自定义事件，`properties` 为结构化对象 |

SDK 会自动处理：

- 访客 ID / 会话 ID 的生成与本地持久化（`localStorage`，键名 `wa:<siteId>:visitor_id` / `session_id`）
- 当前页面信息（URL、路径、标题、来源 referrer）
- UTM 参数（`utm_source` / `utm_medium` / `utm_campaign` / `utm_term` / `utm_content`）
- 设备信息（User-Agent、语言）

#### 采集接口（直接调用）

如果不使用 SDK，也可以直接向 Collector 发送 JSON：

```bash
curl -X POST http://localhost:8080/collect \
  -H "Content-Type: application/json" \
  -d '{
    "type": "page_view",
    "site_id": "site_1",
    "public_key": "pk_xxx",
    "occurred_at": "2026-06-26T10:00:00Z",
    "visitor": { "id": "v_abc", "session_id": "s_abc" },
    "page": { "url": "https://example.com/", "path": "/", "title": "Home", "referrer": "" },
    "campaign": { "source": "", "medium": "", "campaign": "" },
    "device": { "user_agent": "Mozilla/5.0", "language": "zh-CN" }
  }'
```

成功返回 `204 No Content`；凭证无效返回 `401`。

---

### 看板 API

Dashboard 同时托管静态页面（`/`）和 JSON API（`/api/...`）。

| 方法 / 路径 | 说明 |
| --- | --- |
| `POST /api/sites` | 创建站点（需 `X-Owner-User-ID` 请求头） |
| `GET /api/sites/{siteID}/online` | 实时在线访客数 |
| `GET /api/sites/{siteID}/overview` | 核心指标汇总 |
| `GET /api/sites/{siteID}/trend` | 指标趋势（折线图数据） |
| `GET /api/sites/{siteID}/dimensions` | 维度报表 |

通用查询参数：

- `grain`：`minute` / `hour` / `day`（默认 `hour`）
- `from` / `to`：时间范围，支持 `2006-01-02` 或 RFC3339
- `dimension`（仅 `/dimensions`）：`page` / `referrer` / `utm` / `device` / `geo` / `event`（默认 `page`）
- `limit`（仅 `/dimensions`）：返回行数上限

示例：

```bash
curl -s "http://localhost:8081/api/sites/site_1/overview?grain=day&from=2026-06-01&to=2026-06-26"
curl -s "http://localhost:8081/api/sites/site_1/dimensions?dimension=referrer&limit=20"
```

---

### 配置

运行时配置通过环境变量读取，参考 [.env.example](.env.example)；YAML 形式的本地参考配置见 [config/local.example.yaml](config/local.example.yaml)。

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `WA_COLLECTOR_ADDR` | `:8080` | Collector 监听地址 |
| `WA_DASHBOARD_ADDR` | `:8081` | Dashboard 监听地址 |
| `WA_POSTGRES_DSN` | `postgres://...localhost:5432/webanalytics` | PostgreSQL 连接串 |
| `WA_REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `WA_REDIS_DB` | `0` | Redis DB |
| `WA_QUEUE_DRIVER` | `redis-streams` | 队列驱动 |
| `WA_QUEUE_STREAM` | `analytics:events` | Stream 名称 |
| `WA_QUEUE_CONSUMER_GROUP` | `analytics-workers` | 消费者组 |
| `WA_STORAGE_DRIVER` | `postgres` | 存储驱动 |
| `WA_WORKER_BATCH_SIZE` | `1000` | Worker 批量写入大小 |
| `WA_WORKER_FLUSH_INTERVAL` | `1s` | Worker 刷新间隔 |
| `WA_ONLINE_WINDOW` | `5m` | 在线访客时间窗口 |
| `WA_RAW_RETENTION_DAYS` | `30` | 原始事件保留天数 |
| `WA_MINUTE_AGG_RETENTION_DAYS` | `15` | 分钟聚合保留天数 |
| `WA_HOUR_AGG_RETENTION_DAYS` | `365` | 小时聚合保留天数 |
| `WA_DAY_AGG_RETENTION_DAYS` | `0` | 天聚合保留天数（0 表示不限） |
| `WA_CREATE_PARTITIONS_AHEAD_DAYS` | `7` | 提前创建分区天数 |

---

### 测试

```bash
# Go 测试
GOCACHE=$(pwd)/.cache/go-build go test ./...

# SDK 测试
npm run test:sdk
```

---

### 目录结构

```text
cmd/                  服务入口（collector / dashboard / worker）
internal/
  collector/          采集服务与处理器
  dashboard/          看板服务、API 处理器、静态页面（static/）
  worker/             事件消费与聚合
  store/              存储接口（EventStore / StatsStore）
  storage/postgres/   PostgreSQL 适配器
  queue/              队列接口与 Redis Streams 实现
  domain/             事件领域模型
  config/             环境变量配置加载
  retention/          数据保留策略规划器
migrations/           数据库迁移 SQL
sdk/                  JavaScript 埋点 SDK
config/               YAML 参考配置
docs/                 运维文档（runbook / observability）
tools/                压测脚本
```

---

### 后续扩展方向

实现保留了清晰的扩展出口：

- **队列**：当前 Redis Streams，后续可通过 `EventPublisher` / `EventConsumer` 切换到 RabbitMQ 或 Kafka
- **分析存储**：当前 PostgreSQL，后续可通过 `EventStore` / `StatsStore` 接入 ClickHouse

即使引入 ClickHouse 作为分析事件存储，PostgreSQL 仍建议保留为用户、站点、凭证、权限和配置的系统记录库。

运维参考：[docs/runbook.md](docs/runbook.md)、[docs/observability.md](docs/observability.md)、[tools/loadtest.mjs](tools/loadtest.mjs)。

---

## English

Self-hosted web traffic analytics covering the full path from frontend tracking to dashboard reporting.

### Components

- **JavaScript SDK** (`sdk/tracker.mjs`): pageview, heartbeat, and custom-event collection
- **Collector** (`cmd/collector`): validates credentials, rate-limits, publishes to the queue, marks online visitors
- **Worker** (`cmd/worker`): batch-consumes events into raw partitions and aggregate tables
- **Dashboard** (`cmd/dashboard`): serves JSON stats APIs plus a browser dashboard
- **Redis Streams**: event queue / buffering
- **Redis online state**: real-time active visitors
- **PostgreSQL**: business data, raw-event partitions, aggregate tables

Capacity target: 30M events/day stable, 50M events/day stretch.

### Data flow

```text
JS SDK -> Collector (POST /collect) -> Redis Streams -> Workers -> PostgreSQL (raw partitions + aggregates)
                  └-> Redis online state
Dashboard reads Redis online state + aggregate tables (not raw partitions).
```

### Ports

| Service | `go run` default | docker compose | Env var |
| --- | --- | --- | --- |
| Collector | `:8080` | `:8085` | `WA_COLLECTOR_ADDR` |
| Dashboard | `:8081` | `:8086` | `WA_DASHBOARD_ADDR` |
| PostgreSQL | `localhost:5432` | `localhost:5432` | `WA_POSTGRES_DSN` |
| Redis | `localhost:6379` | `localhost:6379` | `WA_REDIS_ADDR` |

> docker-compose remaps the HTTP ports to 8085 / 8086 via environment variables. Plain `go run` uses 8080 / 8081.

### Getting started

```bash
# 1. dependencies
docker compose up -d postgres redis

# 2. migrations
for f in migrations/*.sql; do
  psql "postgres://webanalytics:webanalytics@localhost:5432/webanalytics?sslmode=disable" -f "$f"
done

# 3. services
go run ./cmd/collector   # :8080
go run ./cmd/dashboard   # :8081
go run ./cmd/worker

# health
curl -s http://localhost:8080/healthz
curl -s http://localhost:8081/healthz
```

Open the dashboard at `http://localhost:8081/` (or `:8086` under docker compose), enter a site ID, pick a grain and date range, and refresh.

### Tracking / instrumentation

```html
<script type="module">
  import { createTracker } from "/sdk/tracker.mjs";

  const tracker = createTracker({
    siteId: "site_1",
    publicKey: "pk_xxx",
    collectUrl: "http://localhost:8080/collect"
  });

  tracker.trackPageView();
  setInterval(() => tracker.heartbeat(), 30_000);
  tracker.track("signup", { plan: "pro", price: 99 });
</script>
```

| Method | Event type | Notes |
| --- | --- | --- |
| `trackPageView()` | `page_view` | current pageview |
| `heartbeat()` | `heartbeat` | keeps online-visitor state alive |
| `track(name, properties)` | `custom` | custom event with structured properties |

The SDK auto-manages visitor/session IDs (persisted in `localStorage`), page info, UTM params, and device metadata. To bypass the SDK, POST JSON to `/collect` directly (returns `204` on success, `401` on invalid credential).

### Dashboard API

The dashboard service hosts both the static UI (`/`) and JSON APIs (`/api/...`).

| Method / Path | Purpose |
| --- | --- |
| `POST /api/sites` | create a site (requires `X-Owner-User-ID` header) |
| `GET /api/sites/{siteID}/online` | real-time online visitors |
| `GET /api/sites/{siteID}/overview` | core metric totals |
| `GET /api/sites/{siteID}/trend` | trend series |
| `GET /api/sites/{siteID}/dimensions` | dimension report |

Query params: `grain` (`minute`/`hour`/`day`), `from`/`to` (`2006-01-02` or RFC3339), `dimension` (`page`/`referrer`/`utm`/`device`/`geo`/`event`), `limit`.

### Tests

```bash
GOCACHE=$(pwd)/.cache/go-build go test ./...
npm run test:sdk
```

### Status

This is a tested first-implementation skeleton. The collector/dashboard entrypoints currently use in-memory credentials and empty readers; PostgreSQL and Redis adapters exist behind interfaces and wiring to real clients is still being completed. The dashboard frontend calls the real APIs, so displayed data depends on how far the backend readers are wired.

### Future backend exits

- Queue: Redis Streams now, RabbitMQ or Kafka later via `EventPublisher` / `EventConsumer`.
- Analytics store: PostgreSQL now, ClickHouse later via `EventStore` / `StatsStore`.

PostgreSQL should remain the system of record for users, sites, credentials, permissions, and configuration even if ClickHouse becomes the analytical event store.