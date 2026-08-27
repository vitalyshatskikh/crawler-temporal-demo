CREATE TABLE parsing_configs (
    id BIGSERIAL PRIMARY KEY,
    source_id TEXT NOT NULL,
    doc_type TEXT NOT NULL,
    name TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_parsing_configs_source_id ON parsing_configs (source_id);

CREATE UNIQUE INDEX uq_parsing_configs_source_doc_type ON parsing_configs (source_id, doc_type);

CREATE TABLE documents (
    sdoc_id TEXT NOT NULL,
    source_id TEXT NOT NULL,
    doc_type TEXT NOT NULL,
    external_url TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    update_interval_sec INTEGER NOT NULL DEFAULT 86400,
    CONSTRAINT pk_documents PRIMARY KEY (sdoc_id, source_id, doc_type),
    CONSTRAINT ck_documents_updated_at CHECK (updated_at >= created_at)
);

CREATE INDEX idx_documents_source_id ON documents (source_id);
CREATE INDEX idx_documents_doc_type ON documents (doc_type);
