package gateway

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

type RepositoryInterface interface {
	GetBySiteGatewayID(ctx context.Context, numericID int64) (*Gateway, error)
	UpsertMeshMapping(ctx context.Context, mg MeshMapping) error
	DeleteGatewayByEui(ctx context.Context, eui string) error
	UpsertGateway(ctx context.Context, g Gateway) error
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type gatewayRow struct {
	ID            string    `db:"id"`
	EUI           string    `db:"eui"`
	GatewayID     string    `db:"gateway_id"`
	SiteGatewayID int64     `db:"site_gateway_id"`
	Kind          string    `db:"kind"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type gatewaMappingRow struct {
	MeshGatewayEui   string    `db:"mg_eui"`
	BorderGatewayEui string    `db:"bg_eui"`
	UpdatedAt        time.Time `db:"updated_at"`
}

func (r *Repository) GetBySiteGatewayID(ctx context.Context, siteGatewayID int64) (*Gateway, error) {

	var row gatewayRow
	err := r.db.GetContext(ctx, &row, `
		SELECT * FROM gateways WHERE site_gateway_id = $1
	`, siteGatewayID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return toGatewayModel(row), nil
}

func (r *Repository) UpsertMeshMapping(ctx context.Context, mg MeshMapping) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO gateway_mesh_mapping (mg_eui, bg_eui)
		VALUES ($1, $2)
		ON CONFLICT (mg_eui) DO UPDATE
			SET bg_eui = EXCLUDED.bg_eui,
			    updated_at  = now()
	`, mg.MGEUI, mg.BGEUI)
	return err
}

func (r *Repository) DeleteGatewayByEui(ctx context.Context, eui string) error {
	var row gatewayRow
	err := r.db.GetContext(ctx, &row, `
		DELETE FROM gateways WHERE eui = $1
	`, eui)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) UpsertGateway(ctx context.Context, g Gateway) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO gateways (eui, gateway_id, site_gateway_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (eui) DO UPDATE
			SET gateway_id = EXCLUDED.gateway_id,
			SET site_gateway_id = EXCLUDED.site_gateway_id
	`, g.EUI, g.GatewayID, g.SiteGatewayID)
	return err
}

// ── Mappers ───────────────────────────────────────────────────

func toGatewayModel(row gatewayRow) *Gateway {
	return &Gateway{
		ID:            row.ID,
		EUI:           row.EUI,
		GatewayID:     row.GatewayID,
		SiteGatewayID: row.SiteGatewayID,
		Kind:          Kind(row.Kind),
		CreatedAt:     row.CreatedAt,
	}
}
