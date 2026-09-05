ALTER TABLE cameras
    DROP COLUMN stream_protocol,
    DROP COLUMN resolution,
    DROP COLUMN fps,
    DROP COLUMN codec,
    ADD COLUMN rtsp_transport VARCHAR(16) NOT NULL DEFAULT 'automatic' CHECK (rtsp_transport IN ('automatic', 'tcp', 'udp')),
    ADD COLUMN source_on_demand BOOLEAN NOT NULL DEFAULT false;
