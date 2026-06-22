CREATE TABLE IF NOT EXISTS raw_events (
    id text NOT NULL,
    site_id uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    event_type text NOT NULL,
    event_name text,
    event_time timestamptz NOT NULL,
    event_date date NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    visitor_id text NOT NULL,
    session_id text NOT NULL,
    page_url text,
    page_path text,
    page_title text,
    referrer text,
    utm_source text,
    utm_medium text,
    utm_campaign text,
    utm_term text,
    utm_content text,
    browser text,
    os text,
    device_type text,
    country text,
    region text,
    city text,
    ip_hash text,
    user_agent text,
    properties jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (event_date, site_id, id)
) PARTITION BY RANGE (event_date);

CREATE INDEX IF NOT EXISTS idx_raw_events_site_time
    ON raw_events (site_id, event_time DESC);

CREATE INDEX IF NOT EXISTS idx_raw_events_site_visitor_time
    ON raw_events (site_id, visitor_id, event_time DESC);

CREATE INDEX IF NOT EXISTS brin_raw_events_event_time
    ON raw_events USING brin (event_time);

CREATE TABLE IF NOT EXISTS raw_events_default
    PARTITION OF raw_events DEFAULT;
