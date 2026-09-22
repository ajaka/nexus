CREATE TYPE status_enum AS ENUM ('completed','pending','sending','failed');

CREATE TABLE IF NOT EXISTS outbox(
    id SERIAL PRIMARY KEY,
    event TEXT NOT NULL,
    payload BYTEA NOT NULL,
    status status_enum NOT NULL DEFAULT 'pending',
    triggered_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    context TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_outbox_created_at ON outbox(created_at);
CREATE INDEX idx_outbox_updated_at ON outbox(updated_at);

CREATE TRIGGER trigger_outbox_updated_at
BEFORE UPDATE ON outbox
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();