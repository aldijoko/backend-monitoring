CREATE TABLE IF NOT EXISTS cameras (
    id SERIAL PRIMARY KEY,
    edge_id INTEGER NOT NULL REFERENCES edges(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    channel INTEGER NOT NULL DEFAULT 1,
    source_url VARCHAR(512) NOT NULL,
    stream_url VARCHAR(512),
    stream_protocol VARCHAR(16) NOT NULL DEFAULT 'hls' CHECK (stream_protocol IN ('hls', 'webrtc', 'mjpeg')),
    resolution VARCHAR(32),
    fps INTEGER,
    codec VARCHAR(32),
    storage_days INTEGER NOT NULL DEFAULT 30,
    status VARCHAR(16) NOT NULL DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'recording', 'error')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cameras_edge_id ON cameras(edge_id);
