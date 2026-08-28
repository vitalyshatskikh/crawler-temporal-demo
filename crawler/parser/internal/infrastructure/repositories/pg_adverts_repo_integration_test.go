//go:build integration

package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/infrastructure/repositories/testutil"
)

func insertDocument(t *testing.T, pool *pgxpool.Pool, doc domain.Document) {
	_, err := pool.Exec(context.Background(),
		`INSERT INTO documents (sdoc_id, source_id, doc_type, external_url, content_url, body, created_at, updated_at, update_interval_sec)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		string(doc.SdocID), string(doc.SourceID), string(doc.Type), doc.ExternalURL, doc.ContentURL, string(doc.Body),
		pgtype.Timestamptz{Time: doc.CreatedAt, Valid: true},
		pgtype.Timestamptz{Time: doc.UpdatedAt, Valid: true},
		doc.UpdateIntervalSec)
	require.NoError(t, err)
}

func cleanupDocument(t *testing.T, pool *pgxpool.Pool, sdocID, sourceID, docType string) {
	_, _ = pool.Exec(context.Background(),
		`DELETE FROM documents WHERE sdoc_id = $1 AND source_id = $2 AND doc_type = $3`,
		sdocID, sourceID, docType)
}

func TestPGAdvertsRepo_GetDocument_WhenFound(t *testing.T) {
	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "adv-get-found-001",
			SourceID:          "siteapi",
			Type:              domain.DocumentTypeSearchPage,
			ExternalURL:       "https://example.com",
			UpdateIntervalSec: 3600,
			CreatedAt:         time.Now().Add(-24 * time.Hour),
			UpdatedAt:         time.Now(),
		},
		Body: []byte(`{"title":"test"}`),
	}
	insertDocument(t, testutil.TestPool, doc)
	defer cleanupDocument(t, testutil.TestPool, "adv-get-found-001", "siteapi", string(domain.DocumentTypeSearchPage))

	repo := NewPGAdvertsRepo(testutil.TestPool)
	result, err := repo.GetDocument(context.Background(), "adv-get-found-001", "siteapi", domain.DocumentTypeSearchPage)

	require.NoError(t, err)
	assert.Equal(t, doc.SdocID, result.SdocID)
	assert.Equal(t, doc.SourceID, result.SourceID)
	assert.Equal(t, doc.Type, result.Type)
	assert.Equal(t, doc.ExternalURL, result.ExternalURL)
	assert.Equal(t, doc.Body, result.Body)
	assert.Equal(t, doc.UpdateIntervalSec, result.UpdateIntervalSec)
}

func TestPGAdvertsRepo_GetDocument_WhenNotFound(t *testing.T) {
	repo := NewPGAdvertsRepo(testutil.TestPool)
	_, err := repo.GetDocument(context.Background(), "notfound-xyz", "siteapi", domain.DocumentTypeSearchPage)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestPGAdvertsRepo_SaveDocument_WhenNew(t *testing.T) {
	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "adv-save-new-001",
			SourceID:          "siteapi",
			Type:              domain.DocumentTypeSurfedAdvert,
			ExternalURL:       "https://example.com/page1",
			UpdateIntervalSec: 86400,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		Body: []byte(`{"content":"new page"}`),
	}
	defer cleanupDocument(t, testutil.TestPool, "adv-save-new-001", "siteapi", string(domain.DocumentTypeSurfedAdvert))

	repo := NewPGAdvertsRepo(testutil.TestPool)
	err := repo.SaveDocument(context.Background(), doc)
	require.NoError(t, err)

	result, err := repo.GetDocument(context.Background(), "adv-save-new-001", "siteapi", domain.DocumentTypeSurfedAdvert)
	require.NoError(t, err)
	assert.Equal(t, doc.SdocID, result.SdocID)
	assert.Equal(t, doc.Body, result.Body)
	assert.Equal(t, doc.ExternalURL, result.ExternalURL)
}

func TestPGAdvertsRepo_SaveDocument_WhenExisting(t *testing.T) {
	createdAt := time.Now().Add(-48 * time.Hour)
	updatedAt := time.Now().Add(-24 * time.Hour)

	original := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "adv-save-exist-001",
			SourceID:          "siteapi",
			Type:              domain.DocumentTypeDownloadedAdvert,
			ExternalURL:       "https://example.com/original",
			UpdateIntervalSec: 3600,
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
		},
		Body: []byte(`{"original":"body"}`),
	}
	insertDocument(t, testutil.TestPool, original)
	defer cleanupDocument(t, testutil.TestPool, "adv-save-exist-001", "siteapi", string(domain.DocumentTypeDownloadedAdvert))

	updated := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "adv-save-exist-001",
			SourceID:          "siteapi",
			Type:              domain.DocumentTypeDownloadedAdvert,
			ExternalURL:       "https://example.com/updated",
			UpdateIntervalSec: 7200,
			CreatedAt:         createdAt,
			UpdatedAt:         time.Now(),
		},
		Body: []byte(`{"updated":"body"}`),
	}

	repo := NewPGAdvertsRepo(testutil.TestPool)
	err := repo.SaveDocument(context.Background(), updated)
	require.NoError(t, err)

	result, err := repo.GetDocument(context.Background(), "adv-save-exist-001", "siteapi", domain.DocumentTypeDownloadedAdvert)
	require.NoError(t, err)
	assert.Equal(t, "adv-save-exist-001", string(result.SdocID))
	assert.Equal(t, original.CreatedAt.Unix(), result.CreatedAt.Unix())
	assert.Equal(t, updated.UpdatedAt.Unix(), result.UpdatedAt.Unix())
	assert.Equal(t, updated.ExternalURL, result.ExternalURL)
	assert.Equal(t, updated.Body, result.Body)
	assert.Equal(t, updated.UpdateIntervalSec, result.UpdateIntervalSec)
}

func TestPGAdvertsRepo_SaveDocument_WhenOnlyBodyDiffers(t *testing.T) {
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now().Add(-12 * time.Hour)

	original := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "adv-body-diff-001",
			SourceID:          "siteapi",
			Type:              domain.DocumentTypeParsedAdvert,
			ExternalURL:       "https://example.com/page",
			UpdateIntervalSec: 3600,
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
		},
		Body: []byte(`{"old":"content"}`),
	}
	insertDocument(t, testutil.TestPool, original)
	defer cleanupDocument(t, testutil.TestPool, "adv-body-diff-001", "siteapi", string(domain.DocumentTypeParsedAdvert))

	updated := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "adv-body-diff-001",
			SourceID:          "siteapi",
			Type:              domain.DocumentTypeParsedAdvert,
			ExternalURL:       "https://example.com/page",
			UpdateIntervalSec: 3600,
			CreatedAt:         createdAt,
			UpdatedAt:         time.Now(),
		},
		Body: []byte(`{"new":"content"}`),
	}

	repo := NewPGAdvertsRepo(testutil.TestPool)
	err := repo.SaveDocument(context.Background(), updated)
	require.NoError(t, err)

	result, err := repo.GetDocument(context.Background(), "adv-body-diff-001", "siteapi", domain.DocumentTypeParsedAdvert)
	require.NoError(t, err)
	assert.Equal(t, original.CreatedAt.Unix(), result.CreatedAt.Unix())
	assert.Equal(t, updated.Body, result.Body)
}
