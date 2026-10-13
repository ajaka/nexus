CREATE TYPE user_type as ENUM('buyer','seller','both');

CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    middle_name TEXT,
    phone_number TEXT,
    type user_type NOT NULL,
    country TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(first_name, last_name, middle_name, global_id)
);
