## Why

The project exposes dashboard APIs but has no user-facing dashboard page, so operators cannot inspect traffic metrics without manually calling JSON endpoints. A lightweight frontend hosted by the Go dashboard service makes the existing analytics pipeline visible and creates a concrete contract for the reporting fields the API must return.

## What Changes

- Add a static dashboard frontend served by the Go dashboard process.
- Display core traffic metrics: PV, IP count, sessions, active UV, summed active UV, cumulative UV, and blended UV.
- Show trend charts for PV, IP, sessions, UV, active UV, cumulative UV, and blended UV over the selected grain.
- Show referrer domain and referrer page tables with PV, IP, UV, sessions, and event counts where available.
- Extend dashboard API results and aggregate models with IP and derived visitor metrics needed by the frontend.
- Implement PostgreSQL aggregate query methods used by the dashboard APIs.

## Capabilities

### New Capabilities
- `dashboard-frontend`: Browser-based traffic dashboard, hosted by the dashboard service, covering key metrics, trend charts, and referrer reports.

### Modified Capabilities
- `web-traffic-analytics`: Dashboard reporting APIs return the additional metrics and dimensions needed by the frontend.

## Impact

- Affected code: `cmd/dashboard`, `internal/dashboard`, `internal/store`, `internal/storage/postgres`, static dashboard assets, migrations/tests as needed.
- APIs: dashboard JSON responses gain additive fields for IP counts and derived UV metrics.
- Dependencies: the frontend uses a CDN-hosted chart library to avoid adding a local build pipeline.
- Systems: dashboard service serves both `/api/...` JSON endpoints and the static application shell.
