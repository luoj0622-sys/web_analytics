## 1. Dashboard API and Store Metrics

- [x] 1.1 Extend aggregate domain structs and dashboard DTOs with IP count and derived visitor metric fields.
- [x] 1.2 Populate IP counts in worker aggregation for site and dimension stats.
- [x] 1.3 Implement PostgreSQL `QuerySiteStats` and `QueryDimensionStats` against aggregate tables.
- [x] 1.4 Update dashboard service tests and storage tests for the added fields and query behavior.

## 2. Hosted Frontend

- [x] 2.1 Add static dashboard HTML, CSS, and JavaScript assets.
- [x] 2.2 Wire `cmd/dashboard` to serve the static frontend while preserving `/api/` routes.
- [x] 2.3 Render metric cards, traffic line chart, referrer domain table, and referrer page table from existing APIs.

## 3. Verification

- [x] 3.1 Run Go tests.
- [x] 3.2 Run SDK tests if JavaScript package behavior is affected.
- [x] 3.3 Manually smoke-test the dashboard page with the local dashboard server.
