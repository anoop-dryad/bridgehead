package gateway

import "errors"

var (
	ErrNotFound        = errors.New("gateway not found")
	ErrMappingNotFound = errors.New("mesh-border mapping not found")
)
