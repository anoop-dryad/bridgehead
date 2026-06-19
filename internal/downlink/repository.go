// internal/downlink/repository.go
package downlink

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

// db struct lives next to queries — never exposed outside this file
type downlinkRow struct {
	ID         string    `db:"id"`
	DeviceEUI  string    `db:"device_eui"`
	DeviceType string    `db:"device_type"`
	Payload    []byte    `db:"payload"`
	Type       string    `db:"type"`
	Status     string    `db:"status"`
	RetryCount int       `db:"retry_count"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	ExpiresAt  time.Time `db:"expires_at"`
}

func (r *Repository) Create(ctx context.Context, req CreateRequest) (*DownlinkRequest, error) {
	row := downlinkRow{
		DeviceEUI:  req.DeviceEUI,
		DeviceType: string(req.DeviceType),
		Payload:    req.Payload,
		Type:       string(req.Type),
		Status:     string(StatusPending),
		ExpiresAt:  time.Now().Add(DefaultTTL),
	}

	// caller provided id — use it (idempotency key)
	if req.ID != nil {
		row.ID = *req.ID
	}

	// caller provided expiry — use it
	if req.ExpiresAt != nil {
		row.ExpiresAt = *req.ExpiresAt
	}

	query := `
        INSERT INTO downlink_requests
            (id, device_eui, device_type, payload, type, status, expires_at)
        VALUES
            (COALESCE(:id, gen_random_uuid()), :device_eui, :device_type,
             :payload, :type, :status, :expires_at)
        ON CONFLICT (id) DO NOTHING
        RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, query, row)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// ON CONFLICT DO NOTHING returns no rows — means duplicate ID
	if !rows.Next() {
		return nil, ErrDuplicateID
	}

	var result downlinkRow
	if err := rows.StructScan(&result); err != nil {
		return nil, err
	}
	return toModel(result), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*DownlinkRequest, error) {
	var row downlinkRow
	err := r.db.GetContext(ctx, &row, `
        SELECT * FROM downlink_requests WHERE id = $1
    `, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toModel(row), nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status Status) error {
	result, err := r.db.ExecContext(ctx, `
        UPDATE downlink_requests
        SET status     = $1,
            updated_at = now()
        WHERE id = $2
    `, string(status), id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
        DELETE FROM downlink_requests WHERE id = $1
    `, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context, deviceEUI string) ([]*DownlinkRequest, error) {
	var rows []downlinkRow
	err := r.db.SelectContext(ctx, &rows, `
        SELECT * FROM downlink_requests
        WHERE device_eui = $1
        ORDER BY created_at DESC
    `, deviceEUI)
	if err != nil {
		return nil, err
	}
	return toModels(rows), nil
}

func toModel(row downlinkRow) *DownlinkRequest {
	return &DownlinkRequest{
		ID:         row.ID,
		DeviceEUI:  row.DeviceEUI,
		DeviceType: DeviceType(row.DeviceType),
		Payload:    row.Payload,
		Type:       Type(row.Type),
		Status:     Status(row.Status),
		RetryCount: row.RetryCount,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		ExpiresAt:  row.ExpiresAt,
	}
}

func toModels(rows []downlinkRow) []*DownlinkRequest {
	result := make([]*DownlinkRequest, len(rows))
	for i, r := range rows {
		result[i] = toModel(r)
	}
	return result
}
