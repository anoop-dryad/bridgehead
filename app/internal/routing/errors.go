package routing

import "errors"

var (
	ErrUnknownKind = errors.New("unknown kind : available kinds are 'border', 'mesh', 'sensor'")
)
