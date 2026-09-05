CREATE TABLE IF NOT EXISTS edges (
    id SERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255),
    ip_address VARCHAR(64),
    location VARCHAR(255),
    status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('online', 'offline', 'pending')),
    last_heartbeat TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
