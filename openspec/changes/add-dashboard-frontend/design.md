## Overview

Serve a small static dashboard application from the existing Go dashboard service. The page calls the existing dashboard API family, renders metric cards, draws multi-series trend charts, and shows referrer domain/page tables. Backend changes remain aggregate-first and additive.

## Frontend

- Location: `internal/dashboard/static/`.
- Serving: `cmd/dashboard` registers a file server for `/` while keeping `/api/...` on the JSON handler.
- Stack: plain HTML/CSS/JavaScript plus ECharts from CDN.
- Inputs:
  - `site_id` text field.
  - Grain selector: `minute`, `hour`, `day`.
  - Refresh button.
- API calls:
  - `/api/sites/{siteID}/online`
  - `/api/sites/{siteID}/overview?grain={grain}`
  - `/api/sites/{siteID}/trend?grain={grain}`
  - `/api/sites/{siteID}/dimensions?dimension=referrer`
- Display:
  - Metric strip for PV, IP, session, active UV, summed active UV, cumulative UV, blended UV.
  - ECharts line chart with selectable series for PV, IP, session, UV variants.
  - Referrer domain and referrer page tables. Domain rows are computed client-side from referrer URLs returned by the current dimension API; page rows retain full referrer keys.

## Backend Metrics

Additive fields will be introduced so old clients keep working:

- `ip_count`: aggregate count of unique IP hashes in a bucket.
- `active_visitors`: real-time online count for overview.
- `summed_active_visitors`: sum of per-bucket unique visitors across the selected period.
- `cumulative_visitors`: cumulative unique visitors represented by aggregate rows. With the current aggregate-only storage this is approximated as the maximum cumulative running sum available from rows, not a raw-event distinct scan.
- `blended_visitors`: a conservative blended UV value for the selected period, calculated from aggregate rows and active visitor count.

The current storage shape does not preserve exact visitor sets after aggregation, so exact cross-day unique IP/UV cannot be recovered from aggregate tables alone. This change exposes aggregate-safe approximations and leaves the raw-event distinct scan out of the routine dashboard path.

## PostgreSQL Queries

Implement `StatsStore.QuerySiteStats` and `StatsStore.QueryDimensionStats` against aggregate tables:

- Select rows by `site_id`, optional `from`/`to`, and grain/dimension.
- Sort by bucket ascending for trends.
- Limit dimension rows, defaulting to a practical top-N when the query does not specify one.
- Preserve aggregate-first behavior and avoid raw-event scans.

## Risks

- Chart library CDN availability affects local page rendering when offline.
- Exact cross-day de-duplication cannot be provided without storing distinct visitor/IP sketches or querying raw events.
- The existing aggregate writer currently stores hourly site stats and daily dimension stats; day/minute site trends may be empty until writers produce those grains.
