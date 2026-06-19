package downlink

// validTransitions defines allowed state changes
var validTransitions = map[Status][]Status{
	StatusPending:    {StatusQueued, StatusExpired},
	StatusQueued:     {StatusDispatched, StatusExpired},
	StatusDispatched: {StatusDelivered, StatusQueued, StatusFailed},
	StatusDelivered:  {}, // terminal
	StatusFailed:     {}, // terminal
	StatusExpired:    {}, // terminal
}

func (r *DownlinkRequest) CanTransitionTo(next Status) bool {
	allowed, ok := validTransitions[r.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

func (r *DownlinkRequest) TransitionTo(next Status) error {
	if !r.CanTransitionTo(next) {
		return ErrInvalidStatus
	}
	r.Status = next
	return nil
}
