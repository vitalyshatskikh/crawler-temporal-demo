ALTER TABLE adverts ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_adverts_deleted_at ON adverts (deleted_at) WHERE deleted_at IS NOT NULL;