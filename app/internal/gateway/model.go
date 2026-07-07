package gateway

import "time"

type Kind string

const (
	TypeBG Kind = "bg"
	TypeMG Kind = "mg"
)

type Gateway struct {
	ID            string
	EUI           string
	GatewayID     string
	SiteGatewayID int64
	Kind          Kind
	CreatedAt     time.Time
}

type MeshMapping struct {
	BGEUI     string
	MGEUI     string
	UpdatedAt time.Time
}
