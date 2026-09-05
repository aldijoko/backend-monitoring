ALTER TABLE cameras
    DROP COLUMN rtsp_transport,
    DROP COLUMN source_on_demand,
    ADD COLUMN stream_protocol VARCHAR(16) NOT NULL DEFAULT 'hls' CHECK (stream_protocol IN ('hls', 'webrtc', 'mjpeg')),
    ADD COLUMN resolution VARCHAR(32),
    ADD COLUMN fps INTEGER,
    ADD COLUMN codec VARCHAR(32);
