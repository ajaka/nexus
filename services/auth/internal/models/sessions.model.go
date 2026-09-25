package models

import (
	"time"

	"github.com/google/uuid"
)

type Sessions struct {
	Id        uuid.UUID `json:"id" db:"id" redis:"id"`
	UserId    uuid.UUID `json:"user_id" db:"user_id" redis:"user_id"`
	SessionId string    `json:"session_id" db:"session_id" redis:"session_id"`
	Os        string    `json:"os" db:"os" redis:"os"`
	Browser   string    `json:"browser" db:"browser" redis:"browser"`
	Brand     string    `json:"brand" db:"brand" redis:"brand"`
	Ip        string    `json:"ip" db:"ip" redis:"ip"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}
