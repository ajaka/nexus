CREATE TYPE seller_business_type AS ENUM ('sole_proprietorship', 'llc', 'corporation', 'individual');

CREATE TABLE IF NOT EXISTS seller_profile (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    business_name TEXT NOT NULL,
    business_type seller_business_type NOT NULL,
    tax_identifier TEXT UNIQUE,
    is_identity_verified BOOLEAN DEFAULT false NOT NULL,
    verification_submitted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, business_name, business_type, tax_identifier)
);

CREATE TABLE IF NOT EXISTS seller_storefront (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    seller_id UUID REFERENCES seller_profile(user_id) ON DELETE CASCADE NOT NULL,
    store_name TEXT UNIQUE NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    store_description TEXT,
    support_email TEXT NOT NULL,
    return_street_address TEXT NOT NULL,
    return_city TEXT NOT NULL,
    return_country TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_seller_storefront_seller_id ON seller_storefront(seller_id);
CREATE INDEX idx_seller_storefront_slug ON seller_storefront(slug);
