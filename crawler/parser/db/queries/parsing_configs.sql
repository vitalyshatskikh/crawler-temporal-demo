-- name: GetParsingConfig :one
SELECT id, source_id, doc_type, name, config, external_url_jmespath, external_url_template, content_url_template
FROM parsing_configs
WHERE source_id = $1 AND doc_type = $2;
