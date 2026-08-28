-- name: GetDocument :one
SELECT sdoc_id, source_id, doc_type, external_url, content_url, body, created_at, updated_at, update_interval_sec
FROM documents
WHERE sdoc_id = $1 AND source_id = $2 AND doc_type = $3;

-- name: UpsertDocument :exec
INSERT INTO documents (sdoc_id, source_id, doc_type, external_url, content_url, body, created_at, updated_at, update_interval_sec)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (sdoc_id, source_id, doc_type) DO UPDATE SET
    external_url = EXCLUDED.external_url,
    content_url = EXCLUDED.content_url,
    body = EXCLUDED.body,
    updated_at = EXCLUDED.updated_at,
    update_interval_sec = EXCLUDED.update_interval_sec;
