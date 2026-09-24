package domain

import (
	"github.com/google/uuid"
)

type MinimalUserStruct struct {
	UserId uuid.UUID `json:"id"`
	Email  string    `json:"email"`
}

type LimiterResponse struct {
	Allowed    bool
	Remaining  uint64
	RetryAfter uint64
}
