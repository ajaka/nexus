package models

import (
	"time"

	"github.com/google/uuid"
)

type Outbox struct {
	Id          int64     `json:"id" db:"id"`
	Event       string    `json:"event" db:"event"`
	Payload     []byte    `json:"payload" db:"payload"`
	TriggeredBy uuid.UUID `json:"triggered_by" db:"triggerred_by"`
	Status      string    `json:"status" db:"status"`
	Context     string    `json:"context" db:"context"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
