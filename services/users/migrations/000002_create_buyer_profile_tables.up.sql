CREATE TABLE IF NOT EXISTS buyer_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    marketing_opt_in BOOLEAN DEFAULT false NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS buyer_address (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    buyer_id UUID REFERENCES buyer_profiles(user_id) ON DELETE CASCADE NOT NULL,
    default_shipping BOOLEAN DEFAULT false NOT NULL,
    default_billing BOOLEAN DEFAULT false NOT NULL,
    street_address_1 TEXT NOT NULL,
    street_address_2 TEXT,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    country_code TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(buyer_id, street_address_1, street_address_2, city, state)
);


CREATE INDEX idx_buyer_default_shipping ON buyer_address(buyer_id) WHERE default_shipping = true;
