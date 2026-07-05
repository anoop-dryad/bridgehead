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
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type GatewayRow struct {
	ID            string    `db:"id"`
	EUI           string    `db:"eui"`
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

	var row GatewayRow
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

// ── Mappers ───────────────────────────────────────────────────

func toGatewayModel(row GatewayRow) *Gateway {
	return &Gateway{
		ID:            row.ID,
		EUI:           row.EUI,
		SiteGatewayID: row.SiteGatewayID,
		Kind:          row.Kind,
		CreatedAt:     row.CreatedAt,
	}
}
