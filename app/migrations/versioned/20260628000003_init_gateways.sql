CREATE TYPE gateway_kind AS ENUM ('bg', 'mg');

CREATE TABLE gateways (
    eui             VARCHAR(32) PRIMARY KEY,
    site_gateway_id BIGINT      NOT NULL UNIQUE,
    kind            gateway_kind NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_gateways_site_gateway_id ON gateways(site_gateway_id);
CREATE INDEX idx_gateways_kind       ON gateways(kind);

CREATE TABLE gateway_mesh_mapping (
    bg_eui     VARCHAR(32) PRIMARY KEY REFERENCES gateways(eui),
    mg_eui     VARCHAR(32) NOT NULL    REFERENCES gateways(eui),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_gmm_mg_eui ON gateway_mesh_mapping(mg_eui);