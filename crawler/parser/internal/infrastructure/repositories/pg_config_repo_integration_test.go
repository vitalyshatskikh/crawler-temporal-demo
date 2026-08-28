//go:build integration

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/infrastructure/repositories/testutil"
)

func insertParsingConfig(t *testing.T, pool *pgxpool.Pool, sourceID, docType, name string, paramsJSON string) {
	_, err := pool.Exec(context.Background(),
		`INSERT INTO parsing_configs (source_id, doc_type, name, config) VALUES ($1, $2, $3, $4::jsonb)`,
		sourceID, docType, name, paramsJSON)
	require.NoError(t, err)
}

func insertParsingConfigWithURLFields(t *testing.T, pool *pgxpool.Pool, sourceID, docType, name string, paramsJSON, externalURLJMESPath, externalURLTemplate, contentURLTemplate string) {
	_, err := pool.Exec(context.Background(),
		`INSERT INTO parsing_configs (source_id, doc_type, name, config, external_url_jmespath, external_url_template, content_url_template) VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7)`,
		sourceID, docType, name, paramsJSON, externalURLJMESPath, externalURLTemplate, contentURLTemplate)
	require.NoError(t, err)
}

func cleanupParsingConfig(t *testing.T, pool *pgxpool.Pool, sourceID, docType string) {
	_, _ = pool.Exec(context.Background(),
		`DELETE FROM parsing_configs WHERE source_id = $1 AND doc_type = $2`,
		sourceID, docType)
}

func TestPGConfigRepo_GetConfig_WhenRowExists(t *testing.T) {
	params := []domain.ParsingParam{
		{Name: "external_url", JMESPath: "urls[*]", Default: ""},
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	insertParsingConfig(t, testutil.TestPool, "siteapi", "search_page", "search-config", string(paramsJSON))
	defer cleanupParsingConfig(t, testutil.TestPool, "siteapi", "search_page")

	repo := NewPGConfigRepo(testutil.TestPool)
	cfg, err := repo.GetConfig(context.Background(), "siteapi", "search_page")

	require.NoError(t, err)
	assert.Equal(t, "search-config", cfg.Name)
	assert.Equal(t, domain.SourceID("siteapi"), cfg.SourceID)
	assert.Equal(t, domain.DocumentType("search_page"), cfg.DocumentType)
	assert.Equal(t, params, cfg.Params)
}

func TestPGConfigRepo_GetConfig_WhenNoRowForDocType(t *testing.T) {
	params := []domain.ParsingParam{
		{Name: "external_url", JMESPath: "urls[*]", Default: ""},
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	insertParsingConfig(t, testutil.TestPool, "siteapi", "other_type", "other-config", string(paramsJSON))
	defer cleanupParsingConfig(t, testutil.TestPool, "siteapi", "other_type")

	repo := NewPGConfigRepo(testutil.TestPool)
	_, err = repo.GetConfig(context.Background(), "siteapi", "search_page")

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestPGConfigRepo_GetConfig_WhenSameSourceDifferentDocType(t *testing.T) {
	params1 := []domain.ParsingParam{{Name: "url1", JMESPath: "path1", Default: ""}}
	params2 := []domain.ParsingParam{{Name: "url2", JMESPath: "path2", Default: ""}}
	paramsJSON1, _ := json.Marshal(params1)
	paramsJSON2, _ := json.Marshal(params2)

	insertParsingConfig(t, testutil.TestPool, "siteapi", "search_page", "search-config", string(paramsJSON1))
	defer cleanupParsingConfig(t, testutil.TestPool, "siteapi", "search_page")
	insertParsingConfig(t, testutil.TestPool, "siteapi", "advert_page", "advert-config", string(paramsJSON2))
	defer cleanupParsingConfig(t, testutil.TestPool, "siteapi", "advert_page")

	repo := NewPGConfigRepo(testutil.TestPool)

	cfg, err := repo.GetConfig(context.Background(), "siteapi", "search_page")
	require.NoError(t, err)
	assert.Equal(t, domain.DocumentType("search_page"), cfg.DocumentType)
	assert.Equal(t, "search-config", cfg.Name)

	cfg, err = repo.GetConfig(context.Background(), "siteapi", "advert_page")
	require.NoError(t, err)
	assert.Equal(t, domain.DocumentType("advert_page"), cfg.DocumentType)
	assert.Equal(t, "advert-config", cfg.Name)
}

func TestPGConfigRepo_GetConfig_WhenMalformedConfigJSON(t *testing.T) {
	insertParsingConfig(t, testutil.TestPool, "siteapi", "search_page", "bad-config", `"not an array"`)
	defer cleanupParsingConfig(t, testutil.TestPool, "siteapi", "search_page")

	repo := NewPGConfigRepo(testutil.TestPool)
	_, err := repo.GetConfig(context.Background(), "siteapi", "search_page")

	assert.ErrorIs(t, err, domain.ErrParsingFailed)
	assert.Error(t, err)
}

func TestPGConfigRepo_GetConfig_WhenEmptyParamsArray(t *testing.T) {
	params := []domain.ParsingParam{}
	paramsJSON, _ := json.Marshal(params)

	insertParsingConfig(t, testutil.TestPool, "siteapi", "search_page", "empty-config", string(paramsJSON))
	defer cleanupParsingConfig(t, testutil.TestPool, "siteapi", "search_page")

	repo := NewPGConfigRepo(testutil.TestPool)
	cfg, err := repo.GetConfig(context.Background(), "siteapi", "search_page")

	require.NoError(t, err)
	assert.NotNil(t, cfg.Params)
	assert.Empty(t, cfg.Params)
}

func TestPGConfigRepo_GetConfig_WhenURLFieldsSet_ThenRoundTrip(t *testing.T) {
	params := []domain.ParsingParam{
		{Name: "title", JMESPath: "titles", Default: ""},
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	insertParsingConfigWithURLFields(t, testutil.TestPool, "siteapi", "search_page", "url-fields-config", string(paramsJSON), "urls[*]", "{{_external_url}}?ref=search", "https://cdn.example.com{{_external_url}}")
	defer cleanupParsingConfig(t, testutil.TestPool, "siteapi", "search_page")

	repo := NewPGConfigRepo(testutil.TestPool)
	cfg, err := repo.GetConfig(context.Background(), "siteapi", "search_page")

	require.NoError(t, err)
	assert.Equal(t, "url-fields-config", cfg.Name)
	assert.Equal(t, domain.SourceID("siteapi"), cfg.SourceID)
	assert.Equal(t, domain.DocumentType("search_page"), cfg.DocumentType)
	assert.Equal(t, "urls[*]", cfg.ExternalURLJMESPath)
	assert.Equal(t, "{{_external_url}}?ref=search", cfg.ExternalURLTemplate)
	assert.Equal(t, "https://cdn.example.com{{_external_url}}", cfg.ContentURLTemplate)
	assert.Equal(t, params, cfg.Params)
}

var _ = fmt.Sprint
