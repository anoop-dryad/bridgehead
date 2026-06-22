package downlink

import "errors"

var (
	ErrNotFound            = errors.New("downlink request not found")
	ErrDuplicateID         = errors.New("downlink request with this id already exists")
	ErrInvalidStatus       = errors.New("invalid status transition")
	ErrInvalidDeviceType   = errors.New("invalid device type")
	ErrInvalidDownlinkType = errors.New("invalid downlink type")
	ErrExpired             = errors.New("downlink request has expired")
	ErrMaxRetriesExceeded  = errors.New("max retries exceeded")
)
