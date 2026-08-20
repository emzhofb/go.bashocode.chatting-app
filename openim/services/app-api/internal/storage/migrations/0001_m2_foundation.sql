CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    openim_user_id varchar(64) NOT NULL UNIQUE,
    openim_provisioned_at timestamptz,
    username varchar(32) NOT NULL UNIQUE,
    email varchar(320) NOT NULL UNIQUE,
    password_hash text NOT NULL,
    display_name varchar(80) NOT NULL,
    avatar_url text,
    status varchar(16) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    role varchar(16) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    email_verified_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id),
    access_token_hash char(64) NOT NULL UNIQUE,
    refresh_token_hash char(64) NOT NULL UNIQUE,
    access_expires_at timestamptz NOT NULL,
    refresh_expires_at timestamptz NOT NULL,
    user_agent text,
    ip_address inet,
    last_seen_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_refresh_lookup_idx ON sessions (refresh_token_hash, revoked_at);

