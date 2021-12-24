package http

import "time"

type RateLimitError struct {
	PenaltyTicket 		string
	PenaltyTime			int64
	PenaltyExpiration	time.Time
	PenaltyCaptcha 		bool
}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded"
}