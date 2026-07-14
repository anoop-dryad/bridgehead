package downlink

import "time"

const MaxRetries = 5

type Status string

const (
	StatusQueued     Status = "queued"
	StatusDispatched Status = "dispatched"
	StatusFailed     Status = "failed"
	StatusExpired    Status = "expired"
)

// validTransitions defines allowed state changes
var transitions = map[Status][]Status{
	StatusQueued:     {StatusDispatched, StatusFailed, StatusExpired},
	StatusDispatched: {}, // terminal
	StatusFailed:     {}, // terminal
	StatusExpired:    {}, // terminal
}

// CanTransition reports whether from→to is legal.
func CanTransition(from, to Status) bool {
	for _, allowed := range transitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

func IsTerminal(s Status) bool {
	return len(transitions[s]) == 0
}

type DispatchResult int

const (
	DispatchSuccess DispatchResult = iota
	DispatchFailed
)

// Decision — what the machine decides should happen next.
// Separates "what state" from "what side-effects" (retry bump, backoff).
type Decision struct {
	Next           Status
	IncrementRetry bool
	NextEligibleAt time.Time // only meaningful when Next == StatusQueued
}

// Decide is the heart of the machine: given the current request, the outcome
// of a dispatch attempt, and the current time, compute the next state and its
// bookkeeping. Pure function — no I/O, fully unit-testable.
func Decide(req *DownlinkRequest, result DispatchResult, now time.Time) Decision {
	// expiry takes precedence over everything — time is the ultimate arbiter
	if now.After(req.ExpiresAt) {
		return Decision{Next: StatusExpired}
	}

	switch result {
	case DispatchSuccess:
		return Decision{Next: StatusDispatched}

	case DispatchFailed:
		// one more attempt used; have we exhausted the budget?
		if req.RetryCount+1 >= MaxRetries {
			return Decision{Next: StatusFailed}
		}
		return Decision{
			Next:           StatusQueued,
			IncrementRetry: true,
			NextEligibleAt: now.Add(backoff(req.RetryCount + 1)),
		}

	default:
		return Decision{Next: req.Status} // no change
	}
}

// backoff — exponential, capped at 32s: 2,4,8,16,32,32...
// spreads retries across the timescale on which BG availability actually changes.
func backoff(retryCount int) time.Duration {
	d := time.Duration(1<<uint(retryCount)) * time.Second
	if d > 32*time.Second {
		return 32 * time.Second
	}
	return d
}
