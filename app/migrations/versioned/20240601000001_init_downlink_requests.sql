CREATE TYPE downlink_status AS ENUM (
    'pending',
    'queued',
    'dispatched',
    'delivered',
    'failed',
    'expired'
);

CREATE TYPE downlink_type AS ENUM (
    'config',
    'command',
    'firmware',
    'ack'
);

CREATE TYPE device_type AS ENUM (
    'gateway',
    'sensor'
);

CREATE TABLE downlink_requests (
    id          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    device_eui  VARCHAR(64)     NOT NULL,
    device_type device_type     NOT NULL,
    payload     BYTEA           NOT NULL,
    type        downlink_type   NOT NULL,
    status      downlink_status NOT NULL DEFAULT 'pending',
    retry_count INT             NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ     NOT NULL
);

CREATE UNIQUE INDEX idx_downlink_id         ON downlink_requests(id);
CREATE INDEX idx_downlink_device_eui        ON downlink_requests(device_eui);
CREATE INDEX idx_downlink_status            ON downlink_requests(status);
CREATE INDEX idx_downlink_status_expires_at ON downlink_requests(status, expires_at);