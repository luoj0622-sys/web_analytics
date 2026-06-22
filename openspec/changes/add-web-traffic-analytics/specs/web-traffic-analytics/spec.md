## ADDED Requirements

### Requirement: Site tracking integration
The system SHALL allow an owner to create a tracked site and obtain a JavaScript tracking snippet that can be embedded into that site.

#### Scenario: Tracking snippet generated
- **WHEN** an authenticated owner creates a tracked site
- **THEN** the system SHALL generate a site identifier, a collection credential, and an embeddable JavaScript snippet for that site

#### Scenario: Tracking snippet identifies the site
- **WHEN** the JavaScript SDK sends an analytics event
- **THEN** the event SHALL include enough site credential data for the collector to validate the target site

### Requirement: JavaScript event collection
The JavaScript SDK SHALL collect common traffic analytics events, including page views, heartbeats, referrer data, UTM attributes, device/browser metadata, session information, and custom events.

#### Scenario: Page view event sent
- **WHEN** a visitor loads a tracked page
- **THEN** the SDK SHALL send a page view event containing site identity, page URL, referrer, timestamp, visitor identity, and session identity

#### Scenario: Heartbeat updates activity
- **WHEN** a visitor remains active on a tracked page
- **THEN** the SDK SHALL send periodic heartbeat events that can update online-user state without requiring a raw page-view event

#### Scenario: Custom event sent
- **WHEN** the tracked site calls the SDK custom-event API
- **THEN** the SDK SHALL send an event name and optional structured properties to the collector

### Requirement: Collector ingestion path
The collector SHALL validate incoming analytics events, enrich them with server-observed metadata, publish them to the event queue, update real-time online state, and return without synchronously writing raw events to PostgreSQL.

#### Scenario: Valid event accepted
- **WHEN** the collector receives a valid event for an enabled site
- **THEN** it SHALL publish the event to the configured queue, update online activity state when applicable, and return a successful no-content response

#### Scenario: Invalid site rejected
- **WHEN** the collector receives an event with an invalid or disabled site credential
- **THEN** it SHALL reject the event without publishing it to the queue

#### Scenario: PostgreSQL unavailable during collection
- **WHEN** PostgreSQL is temporarily unavailable
- **THEN** the collector SHALL continue accepting valid events as long as the queue and online-state backend are available

### Requirement: Queue-backed event pipeline
The system SHALL use a producer/consumer event pipeline with Redis Streams as the initial queue implementation.

#### Scenario: Event published to Redis Streams
- **WHEN** the collector accepts a valid analytics event
- **THEN** it SHALL publish the event envelope to a Redis Stream for worker consumption

#### Scenario: Worker acknowledges successful batch
- **WHEN** a worker successfully processes a batch of queued events
- **THEN** it SHALL acknowledge those queue messages so they are not processed again

#### Scenario: Worker retries failed batch
- **WHEN** a worker fails to process one or more queued events
- **THEN** the system SHALL retain or reclaim those events for retry according to the queue implementation

#### Scenario: Queue backlog observable
- **WHEN** events are waiting for worker processing
- **THEN** the system SHALL expose queue length, consumer lag, pending message count, and failed-message metrics

### Requirement: Batch event processing
Workers SHALL process queued events in batches, persist raw events in bulk, and aggregate metrics in memory before writing aggregate updates.

#### Scenario: Raw events written in batch
- **WHEN** a worker consumes a batch of events
- **THEN** it SHALL write raw events to PostgreSQL using batch-oriented insertion rather than one transaction per event

#### Scenario: Aggregates compressed before upsert
- **WHEN** a worker computes statistics for a batch of events
- **THEN** it SHALL combine events by metric dimensions before updating aggregate tables

#### Scenario: Worker scale-out
- **WHEN** traffic volume increases
- **THEN** the system SHALL allow additional worker instances to consume from the queue without changing collector behavior

### Requirement: PostgreSQL partitioned raw storage
The system SHALL store raw analytics events in PostgreSQL daily partitions with bounded indexing and automated partition lifecycle management.

#### Scenario: Event stored in daily partition
- **WHEN** a raw event is persisted
- **THEN** it SHALL be written to the partition corresponding to the event date

#### Scenario: Partition created before use
- **WHEN** a new event date approaches
- **THEN** the system SHALL create the needed raw-event partition before writes require it

#### Scenario: Raw indexes are bounded
- **WHEN** raw-event partitions are created
- **THEN** they SHALL include only required indexes for site/time access and time-ordered scans

### Requirement: Aggregate reporting storage
The system SHALL maintain aggregate tables for common reporting dimensions and time grains.

#### Scenario: Site trend available
- **WHEN** events are processed
- **THEN** the system SHALL update site-level minute, hour, and day aggregates for page views, visitors, sessions, and custom-event counts

#### Scenario: Dimension reports available
- **WHEN** events include page, referrer, UTM, device, browser, operating system, geography, or event-name dimensions
- **THEN** the system SHALL update aggregate tables that support reports for those dimensions

#### Scenario: Dashboard reads aggregates
- **WHEN** a dashboard requests historical reports
- **THEN** the system SHALL read aggregate tables by default rather than scanning raw-event partitions

### Requirement: Real-time online users
The system SHALL track real-time online users using Redis activity state.

#### Scenario: Visitor marked online
- **WHEN** a visitor sends a page view or heartbeat event
- **THEN** the system SHALL update the visitor's latest activity timestamp for the corresponding site

#### Scenario: Online count returned
- **WHEN** a dashboard requests current online users for a site
- **THEN** the system SHALL count visitors active within the configured online window

#### Scenario: Expired online activity ignored
- **WHEN** a visitor has not sent activity within the configured online window
- **THEN** the visitor SHALL not be counted as currently online

### Requirement: Hot warm cold retention
The system SHALL apply hot, warm, and cold data retention policies to raw events and aggregates.

#### Scenario: Raw retention enforced
- **WHEN** a raw-event partition exceeds the configured raw retention window
- **THEN** the system SHALL detach, archive, or drop the partition according to retention configuration

#### Scenario: Aggregate retention enforced
- **WHEN** minute, hour, or day aggregate data exceeds its configured retention window
- **THEN** the system SHALL retain or remove it according to the configured grain-specific policy

#### Scenario: Long-term reports remain available
- **WHEN** raw events have expired
- **THEN** long-term trend reports SHALL remain available from retained day-level or month-level aggregates

### Requirement: Capacity targets and degradation
The system SHALL be designed for 30 million events per day as a stable target and 50 million events per day as a stretch target.

#### Scenario: Stable target documented
- **WHEN** the system is deployed for production use
- **THEN** operational documentation SHALL define 30 million events per day as the stable target capacity

#### Scenario: Stretch target documented
- **WHEN** capacity planning is performed
- **THEN** operational documentation SHALL define 50 million events per day as a stretch target that depends on batching, partitioning, retention, and aggregate-only dashboard queries

#### Scenario: Backlog degradation
- **WHEN** queue backlog exceeds configured thresholds
- **THEN** the system SHALL continue accepting valid events while exposing alerts and allowing non-critical reporting freshness to degrade

### Requirement: Future queue and analytics store exits
The system SHALL isolate queue and analytics-storage implementations behind interfaces that preserve future migration paths to RabbitMQ/Kafka and ClickHouse.

#### Scenario: Queue implementation isolated
- **WHEN** collector or worker business logic publishes or consumes events
- **THEN** it SHALL use queue interfaces rather than Redis-specific APIs directly

#### Scenario: Analytics store implementation isolated
- **WHEN** workers write raw events or aggregates
- **THEN** they SHALL use storage interfaces rather than scattering PostgreSQL-specific logic through business processing

#### Scenario: Business data remains separate
- **WHEN** a future ClickHouse analytics store is introduced
- **THEN** PostgreSQL SHALL remain the system of record for users, sites, credentials, permissions, and configuration
