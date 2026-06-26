ALTER TABLE stats_site_minute
    ADD COLUMN IF NOT EXISTS ip_count bigint NOT NULL DEFAULT 0;

ALTER TABLE stats_site_hour
    ADD COLUMN IF NOT EXISTS ip_count bigint NOT NULL DEFAULT 0;

ALTER TABLE stats_site_day
    ADD COLUMN IF NOT EXISTS ip_count bigint NOT NULL DEFAULT 0;

ALTER TABLE stats_page_day
    ADD COLUMN IF NOT EXISTS ip_count bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sessions bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS event_count bigint NOT NULL DEFAULT 0;

ALTER TABLE stats_referrer_day
    ADD COLUMN IF NOT EXISTS ip_count bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sessions bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS event_count bigint NOT NULL DEFAULT 0;

ALTER TABLE stats_utm_day
    ADD COLUMN IF NOT EXISTS ip_count bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sessions bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS event_count bigint NOT NULL DEFAULT 0;

ALTER TABLE stats_device_day
    ADD COLUMN IF NOT EXISTS sessions bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS event_count bigint NOT NULL DEFAULT 0;

ALTER TABLE stats_geo_day
    ADD COLUMN IF NOT EXISTS sessions bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS event_count bigint NOT NULL DEFAULT 0;

ALTER TABLE stats_event_day
    ADD COLUMN IF NOT EXISTS page_views bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ip_count bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sessions bigint NOT NULL DEFAULT 0;
