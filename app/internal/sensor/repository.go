package sensor

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

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type sensorRow struct {
	ID        string    `db:"id"`
	EUI       string    `db:"eui"`
	DeviceID  string    `db:"device_id"`
	AppID     string    `db:"app_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type gatewayMappingRow struct {
	SensorEUI  string    `db:"sensor_eui"`
	GatewayEUI string    `db:"gateway_eui"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func (r *Repository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(ctx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ── Sensor identity ───────────────────────────────────────────

func (r *Repository) UpsertSensor(ctx context.Context, s Sensor) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sensors (eui, device_id, app_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (eui) DO UPDATE
			SET device_id = EXCLUDED.device_id,
			    app_id    = EXCLUDED.app_id
	`, s.EUI, s.DeviceID, s.AppID)
	return err
}

func (r *Repository) GetByEUI(ctx context.Context, eui string) (*Sensor, error) {
	var row sensorRow
	err := r.db.GetContext(ctx, &row, `
		SELECT * FROM sensors WHERE eui = $1
	`, eui)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toSensorModel(row), nil
}

// ── Gateway mapping ───────────────────────────────────────────

func (r *Repository) UpsertMapping(ctx context.Context, m GatewayMapping) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sensor_gateway_mapping (sensor_eui, gateway_eui)
		VALUES ($1, $2)
		ON CONFLICT (sensor_eui) DO UPDATE
			SET gateway_eui = EXCLUDED.gateway_eui,
			    updated_at  = now()
	`, m.SensorEUI, m.GatewayEUI)
	return err
}

func (r *Repository) GetMappingBySensorEUI(ctx context.Context, sensorEUI string) (*GatewayMapping, error) {
	var row gatewayMappingRow
	err := r.db.GetContext(ctx, &row, `
		SELECT * FROM sensor_gateway_mapping WHERE sensor_eui = $1
	`, sensorEUI)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMappingNotFound
	}
	if err != nil {
		return nil, err
	}
	return toMappingModel(row), nil
}

func (r *Repository) GetSensorsByGatewayEUI(ctx context.Context, gatewayEUI string) ([]*Sensor, error) {
	var rows []sensorRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT s.* FROM sensors s
		JOIN sensor_gateway_mapping m ON m.sensor_eui = s.eui
		WHERE m.gateway_eui = $1
	`, gatewayEUI)
	if err != nil {
		return nil, err
	}
	return toSensorModels(rows), nil
}

// ── Mappers ───────────────────────────────────────────────────

func toSensorModel(row sensorRow) *Sensor {
	return &Sensor{
		ID:        row.ID,
		EUI:       row.EUI,
		DeviceID:  row.DeviceID,
		AppID:     row.AppID,
		CreatedAt: row.CreatedAt,
	}
}

func toSensorModels(rows []sensorRow) []*Sensor {
	result := make([]*Sensor, len(rows))
	for i, r := range rows {
		result[i] = toSensorModel(r)
	}
	return result
}

func toMappingModel(row gatewayMappingRow) *GatewayMapping {
	return &GatewayMapping{
		SensorEUI:  row.SensorEUI,
		GatewayEUI: row.GatewayEUI,
		UpdatedAt:  row.UpdatedAt,
	}
}
