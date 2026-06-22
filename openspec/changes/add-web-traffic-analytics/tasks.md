## 1. Project Foundation

- [x] 1.1 Create the Go backend module structure for collector APIs, worker processes, dashboard APIs, configuration, logging, and shared domain models
- [x] 1.2 Add local development configuration for PostgreSQL, Redis, and service environment variables
- [x] 1.3 Define application configuration for queue implementation, storage implementation, retention windows, batch sizes, and online-user window
- [x] 1.4 Add structured logging, request IDs, metrics hooks, and health-check endpoints for collector, worker, and dashboard services

## 2. PostgreSQL Schema and Partitioning

- [x] 2.1 Create PostgreSQL migrations for users, sites, tracking credentials, and site settings
- [x] 2.2 Create the partitioned `raw_events` parent table and daily partition schema with bounded indexes
- [x] 2.3 Create aggregate tables for site minute/hour/day statistics
- [x] 2.4 Create aggregate tables for page, referrer, UTM, device, browser, operating system, geography, and custom-event dimensions
- [x] 2.5 Implement partition creation and retention metadata tables
- [x] 2.6 Add migration tests or schema validation checks for required tables, indexes, and partition behavior

## 3. Queue and Storage Abstractions

- [x] 3.1 Define `EventPublisher` and `EventConsumer` interfaces that support publish, consume batch, acknowledge, retry or reclaim, and lag metrics
- [x] 3.2 Implement Redis Streams publisher and consumer using consumer groups
- [x] 3.3 Define `EventStore` and `StatsStore` interfaces for raw-event writes, aggregate writes, and dashboard reads
- [x] 3.4 Implement PostgreSQL event and statistics stores behind the storage interfaces
- [x] 3.5 Add placeholders or contract tests that keep RabbitMQ/Kafka and ClickHouse implementations possible without changing business logic

## 4. JavaScript SDK and Site Integration

- [x] 4.1 Implement site creation and tracking-snippet generation APIs
- [x] 4.2 Implement the JavaScript SDK bootstrap that initializes site identity, visitor identity, and session identity
- [x] 4.3 Implement automatic page-view collection with URL, referrer, UTM, timestamp, and browser metadata
- [x] 4.4 Implement heartbeat collection for real-time online-user tracking
- [x] 4.5 Implement a custom-event SDK API that sends event names and optional structured properties
- [x] 4.6 Add SDK tests for event payload construction, visitor/session persistence, heartbeat behavior, and custom-event calls

## 5. Collector Service

- [x] 5.1 Implement collector endpoints for page views, heartbeats, and custom events
- [x] 5.2 Validate site credentials and reject disabled or unknown sites before queue publication
- [x] 5.3 Enrich accepted events with server-observed IP-derived, user-agent, and request metadata
- [x] 5.4 Publish accepted event envelopes to Redis Streams without synchronously writing PostgreSQL raw events
- [x] 5.5 Update Redis online-user state for page views and heartbeats
- [x] 5.6 Add rate-limiting or abuse-protection hooks at site and client levels
- [x] 5.7 Add collector tests for valid ingestion, invalid credentials, Redis publication, online updates, and PostgreSQL outage independence

## 6. Worker Processing

- [x] 6.1 Implement worker batch consumption from Redis Streams with configurable batch size and idle timeout
- [x] 6.2 Implement batch raw-event insertion into PostgreSQL daily partitions
- [x] 6.3 Implement in-memory batch aggregation for site, page, referrer, UTM, device, browser, operating system, geography, and custom-event dimensions
- [x] 6.4 Implement batch aggregate upserts through `StatsStore`
- [x] 6.5 Implement successful acknowledgement, failed-message retry, pending-message reclaim, and dead-letter handling
- [x] 6.6 Add worker tests for batch success, partial failure, retry/reclaim behavior, idempotency boundaries, and aggregate correctness

## 7. Dashboard Query APIs

- [x] 7.1 Implement real-time online-user API backed by Redis activity state
- [x] 7.2 Implement overview APIs for PV, UV, sessions, events, bounce-rate inputs, and average-duration inputs from aggregate tables
- [x] 7.3 Implement trend APIs for minute, hour, and day grains
- [x] 7.4 Implement dimension report APIs for pages, referrers, UTM campaigns, devices, browsers, operating systems, geography, and custom events
- [x] 7.5 Ensure dashboard APIs default to aggregate tables and do not scan raw-event partitions for routine reports
- [x] 7.6 Add query tests for aggregate reads, online counts, time ranges, site isolation, and empty-state responses

## 8. Retention and Cold Data Lifecycle

- [x] 8.1 Implement scheduled creation of future daily raw-event partitions
- [x] 8.2 Implement raw-event retention jobs for detach, archive placeholder, or drop actions based on configuration
- [x] 8.3 Implement aggregate retention jobs for minute, hour, and day grains
- [x] 8.4 Add operational commands or admin endpoints to inspect partitions, retention status, and archive failures
- [x] 8.5 Add tests for partition creation, retention cutoff selection, safe partition removal, and long-term aggregate availability

## 9. Observability and Capacity Validation

- [x] 9.1 Expose metrics for collector request rate, queue publish latency, Redis Stream length, consumer lag, pending messages, worker throughput, PostgreSQL write latency, and failed batches
- [x] 9.2 Add alerts or documented thresholds for queue backlog, worker lag, partition creation failure, disk growth, and PostgreSQL write latency
- [x] 9.3 Create load-test tooling for stable 30 million events/day equivalent traffic and stretch 50 million events/day equivalent traffic
- [x] 9.4 Verify high-traffic behavior keeps collector responses fast while allowing dashboard freshness to degrade under backlog
- [x] 9.5 Document operational runbooks for scaling workers, handling Redis backlog, PostgreSQL partition maintenance, and future RabbitMQ/Kafka or ClickHouse migration
