package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id          uuid.UUID `json:"id" db:"id" redis:"id"`
	FirstName   string    `json:"first_name" db:"first_name" redis:"first_name"`
	LastName    string    `json:"last_name" db:"last_name" redis:"last_name"`
	MiddleName  string    `json:"middle_name,omitempty" db:"middle_name" redis:"middle_name"`
	PhoneNumber string    `json:"phone_number,omitempty" db:"phone_number" redis:"phone_number"`
	Type        string    `json:"type" db:"type" redis:"type"`
	Country     string    `json:"country" db:"country" redis:"country"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" redis:"updated_at"`
}
