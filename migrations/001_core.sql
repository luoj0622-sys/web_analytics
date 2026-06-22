CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    password_hash text,
    display_name text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sites (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id uuid NOT NULL REFERENCES users(id),
    name text NOT NULL,
    domain text NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (owner_user_id, domain)
);

CREATE TABLE IF NOT EXISTS tracking_credentials (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    public_key text NOT NULL UNIQUE,
    secret_hash text NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    rotated_at timestamptz
);

CREATE TABLE IF NOT EXISTS site_settings (
    site_id uuid PRIMARY KEY REFERENCES sites(id) ON DELETE CASCADE,
    timezone text NOT NULL DEFAULT 'UTC',
    online_window_seconds integer NOT NULL DEFAULT 300,
    raw_retention_days integer NOT NULL DEFAULT 30,
    minute_retention_days integer NOT NULL DEFAULT 15,
    hour_retention_days integer NOT NULL DEFAULT 365,
    day_retention_days integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
