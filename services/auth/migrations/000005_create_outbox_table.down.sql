DROP INDEX IF EXISTS idx_outbox_created_at;
DROP INDEX IF EXISTS idx_outbox_updated_at;
DROP TABLE IF EXISTS outbox;
DROP TYPE IF EXISTS status_enum;

DROP TRIGGER IF EXISTS trigger_outbox_updated_at ON outbox;