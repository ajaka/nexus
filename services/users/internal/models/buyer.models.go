package models

import (
	"time"

	"github.com/google/uuid"
)

type BuyerProfile struct {
	UserId         uuid.UUID `json:"user_id" db:"user_id" redis:"user_id"`
	MarketingOptIn bool      `json:"marketing_opt_in" db:"marketing_opt_in" redis:"marketing_opt_in"`
	Currency       string    `json:"currency" db:"currency" redis:"currency"`
	CreatedAt      time.Time `json:"created_at" db:"created_at" redis:"created_at"`
}

type BuyerAddress struct {
	Id              uuid.UUID `json:"id" db:"id" redis:"id"`
	BuyerId         uuid.UUID `json:"buyer_id" db:"buyer_id" redis:"buyer_id"`
	DefaultShipping bool      `json:"default_shipping" db:"default_shipping" redis:"default_shipping"`
	DefaultBilling  bool      `json:"default_billing" db:"default_billing" redis:"default_billing"`
	StreetAddress1  string    `json:"street_address_1" db:"street_address_1" redis:"street_address_1"`
	StreetAddress2  string    `json:"street_address_2,omitempty" db:"street_address_2" redis:"street_address_2"`
	City            string    `json:"city" db:"city" redis:"city"`
	State           string    `json:"state" db:"state" redis:"state"`
	PostalCode      string    `json:"postal_code" db:"postal_code" redis:"postal_code"`
	CountryCode     string    `json:"country_code" db:"country_code" redis:"country_code"`
	CreatedAt       time.Time `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at" redis:"updated_at"`
}
