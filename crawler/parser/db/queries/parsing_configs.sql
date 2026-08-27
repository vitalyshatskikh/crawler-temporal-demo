-- name: GetParsingConfig :one
SELECT id, source_id, doc_type, name, config
FROM parsing_configs
WHERE source_id = $1 AND doc_type = $2;
