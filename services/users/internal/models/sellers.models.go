package models

import (
	"time"

	"github.com/google/uuid"
)

type SellerProfile struct {
	UserId                  uuid.UUID  `json:"user_id" db:"user_id" redis:"user_id"`
	BusinessName            string     `json:"business_name" db:"business_name" redis:"business_name"`
	BusinessType            string     `json:"business_type" db:"business_type" redis:"business_type"`
	TaxIdentifier           string     `json:"tax_identifier,omitempty" db:"tax_identifier" redis:"tax_identifier"`
	IsIdentityVerified      bool       `json:"is_identity_verified" db:"is_identity_verified" redis:"is_identity_verified"`
	VerificationSubmittedAt *time.Time `json:"verification_submitted_at,omitempty" db:"verification_submitted_at" redis:"verification_submitted_at"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at" redis:"updated_at"`
}

type SellerStorefront struct {
	Id                  uuid.UUID `json:"id" db:"id" redis:"id"`
	SellerId            uuid.UUID `json:"seller_id" db:"seller_id" redis:"seller_id"`
	StoreName           string    `json:"store_name" db:"store_name" redis:"store_name"`
	Slug                string    `json:"slug" db:"slug" redis:"slug"`
	StoreDescription    string    `json:"store_description,omitempty" db:"store_description" redis:"store_description"`
	SupportEmail        string    `json:"support_email" db:"support_email" redis:"support_email"`
	ReturnStreetAddress string    `json:"return_street_address" db:"return_street_address" redis:"return_street_address"`
	ReturnCity          string    `json:"return_city" db:"return_city" redis:"return_city"`
	ReturnCountry       string    `json:"return_country" db:"return_country" redis:"return_country"`
	CreatedAt           time.Time `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at" redis:"updated_at"`
}
