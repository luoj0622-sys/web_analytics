CREATE TABLE IF NOT EXISTS stats_site_minute (
    site_id uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    bucket timestamptz NOT NULL,
    page_views bigint NOT NULL DEFAULT 0,
    unique_visitors bigint NOT NULL DEFAULT 0,
    sessions bigint NOT NULL DEFAULT 0,
    custom_events bigint NOT NULL DEFAULT 0,
    bounces bigint NOT NULL DEFAULT 0,
    total_duration_seconds bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (site_id, bucket)
);

CREATE TABLE IF NOT EXISTS stats_site_hour (LIKE stats_site_minute INCLUDING ALL);
CREATE TABLE IF NOT EXISTS stats_site_day (LIKE stats_site_minute INCLUDING ALL);

CREATE TABLE IF NOT EXISTS stats_page_day (
    site_id uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    bucket date NOT NULL,
    page_path text NOT NULL,
    page_url text NOT NULL,
    page_views bigint NOT NULL DEFAULT 0,
    unique_visitors bigint NOT NULL DEFAULT 0,
    entrances bigint NOT NULL DEFAULT 0,
    exits bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (site_id, bucket, page_path)
);

CREATE TABLE IF NOT EXISTS stats_referrer_day (
    site_id uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    bucket date NOT NULL,
    referrer text NOT NULL,
    page_views bigint NOT NULL DEFAULT 0,
    unique_visitors bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (site_id, bucket, referrer)
);

CREATE TABLE IF NOT EXISTS stats_utm_day (
    site_id uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    bucket date NOT NULL,
    source text NOT NULL DEFAULT '',
    medium text NOT NULL DEFAULT '',
    campaign text NOT NULL DEFAULT '',
    page_views bigint NOT NULL DEFAULT 0,
    unique_visitors bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (site_id, bucket, source, medium, campaign)
);

CREATE TABLE IF NOT EXISTS stats_device_day (
    site_id uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    bucket date NOT NULL,
    device_type text NOT NULL DEFAULT '',
    browser text NOT NULL DEFAULT '',
    os text NOT NULL DEFAULT '',
    page_views bigint NOT NULL DEFAULT 0,
    unique_visitors bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (site_id, bucket, device_type, browser, os)
);

CREATE TABLE IF NOT EXISTS stats_geo_day (
    site_id uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    bucket date NOT NULL,
    country text NOT NULL DEFAULT '',
    region text NOT NULL DEFAULT '',
    city text NOT NULL DEFAULT '',
    page_views bigint NOT NULL DEFAULT 0,
    unique_visitors bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (site_id, bucket, country, region, city)
);

CREATE TABLE IF NOT EXISTS stats_event_day (
    site_id uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    bucket date NOT NULL,
    event_name text NOT NULL,
    event_count bigint NOT NULL DEFAULT 0,
    unique_visitors bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (site_id, bucket, event_name)
);

CREATE TABLE IF NOT EXISTS partition_lifecycle (
    table_name text PRIMARY KEY,
    partition_date date NOT NULL,
    status text NOT NULL,
    archive_uri text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    error_message text
);
