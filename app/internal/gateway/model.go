package gateway

import "time"

type Type string

const (
	TypeBG Type = "bg"
	TypeMG Type = "mg"
)

type Gateway struct {
	ID            string
	EUI           string
	SiteGatewayID int64
	Kind          string
	CreatedAt     time.Time
}

type MeshMapping struct {
	BGEUI     string
	MGEUI     string
	UpdatedAt time.Time
}
