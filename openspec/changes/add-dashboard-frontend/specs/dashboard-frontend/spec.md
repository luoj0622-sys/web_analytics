## ADDED Requirements

### Requirement: Hosted dashboard frontend
The dashboard service SHALL serve a browser-based frontend without requiring a separate frontend build service.

#### Scenario: Dashboard page is loaded
- **WHEN** a user opens the dashboard service root path
- **THEN** the service SHALL return an HTML application shell for viewing site analytics

#### Scenario: Existing APIs remain available
- **WHEN** a user requests dashboard API paths under `/api/`
- **THEN** the service SHALL route those requests to the JSON API handlers rather than the static file server

### Requirement: Core traffic metrics
The frontend SHALL display the core traffic metrics required for a site overview.

#### Scenario: Overview metrics are displayed
- **WHEN** the frontend loads overview data for a site
- **THEN** it SHALL display PV, IP count, sessions, active UV, summed active UV, cumulative UV, and blended UV

#### Scenario: API returns additive metric fields
- **WHEN** the dashboard overview API returns data
- **THEN** the response SHALL include additive fields for IP count and derived visitor metrics without removing the existing PV, UV, sessions, or custom event fields

### Requirement: Traffic trend chart
The frontend SHALL show traffic metrics over time using a line chart.

#### Scenario: Trend rows are rendered
- **WHEN** the frontend receives trend rows for the selected grain
- **THEN** it SHALL render line series for PV, IP count, sessions, UV, active UV, cumulative UV, and blended UV where data is available

### Requirement: Referrer reports
The frontend SHALL show referrer domain and referrer page reports.

#### Scenario: Referrer data is displayed
- **WHEN** the frontend receives referrer dimension rows
- **THEN** it SHALL display a referrer page table and a referrer domain table with PV, IP, UV, sessions, and event metrics where available

### Requirement: Aggregate-first dashboard reads
Dashboard reporting SHALL read aggregate tables for routine overview, trend, and dimension reports.

#### Scenario: PostgreSQL dashboard query executes
- **WHEN** the dashboard service queries site or dimension statistics
- **THEN** it SHALL use aggregate tables for the requested grain or dimension and SHALL NOT scan raw event partitions for routine reports
