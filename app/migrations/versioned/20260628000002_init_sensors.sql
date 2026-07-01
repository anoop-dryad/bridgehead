-- sensor identity
CREATE TABLE sensors (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    eui        VARCHAR(32)  NOT NULL UNIQUE,
    device_id  VARCHAR(128) NOT NULL,
    app_id     VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_sensors_eui ON sensors(eui);

-- sensor-gateway mapping (separate concern, same domain)
CREATE TABLE sensor_gateway_mapping (
    sensor_eui  VARCHAR(32)  PRIMARY KEY REFERENCES sensors(eui),
    gateway_eui VARCHAR(32)  NOT NULL,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_sgm_gateway_eui ON sensor_gateway_mapping(gateway_eui);