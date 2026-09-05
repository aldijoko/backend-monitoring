CREATE TABLE IF NOT EXISTS recordings (
    id SERIAL PRIMARY KEY,
    edge_id INTEGER NOT NULL REFERENCES edges(id) ON DELETE CASCADE,
    edge_code VARCHAR(32),
    camera_id INTEGER NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
    camera_name VARCHAR(255),
    filename VARCHAR(255) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    storage_path VARCHAR(512),
    duration_seconds INTEGER,
    thumbnail_url VARCHAR(512),
    playback_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_recordings_edge_id ON recordings(edge_id);
CREATE INDEX IF NOT EXISTS idx_recordings_camera_id ON recordings(camera_id);
CREATE INDEX IF NOT EXISTS idx_recordings_started_at ON recordings(started_at);
