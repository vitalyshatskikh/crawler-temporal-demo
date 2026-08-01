DROP INDEX IF EXISTS idx_adverts_deleted_at;
ALTER TABLE adverts DROP COLUMN deleted_at;