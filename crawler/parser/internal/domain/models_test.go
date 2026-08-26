package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

func TestSdocIDForURL_WhenSameURL_ThenSameID(t *testing.T) {
	url := "https://example.com/page"
	id1, err1 := domain.SdocIDForURL(url)
	id2, err2 := domain.SdocIDForURL(url)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, id1, id2)
	assert.Len(t, id1, 32)
}

func TestSdocIDForURL_WhenDifferentURLs_ThenDifferentIDs(t *testing.T) {
	id1, err1 := domain.SdocIDForURL("https://example.com/page1")
	id2, err2 := domain.SdocIDForURL("https://example.com/page2")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, id1, id2)
}

func TestSdocIDForURL_WhenURLsDifferOnlyInTrailingSlash_ThenSameID(t *testing.T) {
	id1, err1 := domain.SdocIDForURL("https://x.com/a/")
	id2, err2 := domain.SdocIDForURL("https://x.com/a")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, id1, id2)
}

func TestSdocIDForURL_WhenURLsDifferOnlyInQueryOrder_ThenSameID(t *testing.T) {
	id1, err1 := domain.SdocIDForURL("https://x.com/search?a=1&b=2")
	id2, err2 := domain.SdocIDForURL("https://x.com/search?b=2&a=1")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, id1, id2)
}

func TestSdocIDForURL_WhenURLsDifferOnlyInFragment_ThenSameID(t *testing.T) {
	id1, err1 := domain.SdocIDForURL("https://x.com/a#section")
	id2, err2 := domain.SdocIDForURL("https://x.com/a")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, id1, id2)
}

func TestSdocIDForURL_WhenInvalidURL_ThenErrValidation(t *testing.T) {
	_, err := domain.SdocIDForURL("not a url")

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestDocumentMeta_WhenValid_ThenValidateReturnsNil(t *testing.T) {
	now := time.Now()
	meta := domain.DocumentMeta{
		SdocID:            "abc123",
		CreatedAt:         now,
		UpdatedAt:         now,
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://example.com",
		UpdateIntervalSec: 86400,
	}

	err := meta.Validate()

	assert.NoError(t, err)
}

func TestDocumentMeta_WhenEmptySourceID_ThenErrValidation(t *testing.T) {
	now := time.Now()
	meta := domain.DocumentMeta{
		SdocID:            "abc123",
		CreatedAt:         now,
		UpdatedAt:         now,
		SourceID:          "",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://example.com",
		UpdateIntervalSec: 86400,
	}

	err := meta.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestDocumentMeta_WhenEmptySdocID_ThenErrValidation(t *testing.T) {
	now := time.Now()
	meta := domain.DocumentMeta{
		SdocID:            "",
		CreatedAt:         now,
		UpdatedAt:         now,
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://example.com",
		UpdateIntervalSec: 86400,
	}

	err := meta.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestDocumentMeta_WhenEmptyExternalURL_ThenErrValidation(t *testing.T) {
	now := time.Now()
	meta := domain.DocumentMeta{
		SdocID:            "abc123",
		CreatedAt:         now,
		UpdatedAt:         now,
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "",
		UpdateIntervalSec: 86400,
	}

	err := meta.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestDocumentMeta_WhenUpdatedAtBeforeCreatedAt_ThenErrValidation(t *testing.T) {
	now := time.Now()
	meta := domain.DocumentMeta{
		SdocID:            "abc123",
		CreatedAt:         now,
		UpdatedAt:         now.Add(-time.Second),
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://example.com",
		UpdateIntervalSec: 86400,
	}

	err := meta.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestSdocIDForURL_WhenUppercaseScheme_ThenLowercased(t *testing.T) {
	id1, err1 := domain.SdocIDForURL("HTTPS://example.com/page")
	id2, err2 := domain.SdocIDForURL("https://example.com/page")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, id1, id2)
}

func TestSdocIDForURL_WhenUppercaseHost_ThenLowercased(t *testing.T) {
	id1, err1 := domain.SdocIDForURL("https://EXAMPLE.COM/page")
	id2, err2 := domain.SdocIDForURL("https://example.com/page")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, id1, id2)
}

func TestSdocIDForURL_WhenEmptyString_ThenErrValidation(t *testing.T) {
	_, err := domain.SdocIDForURL("")

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestDocumentMeta_WhenZeroCreatedAt_ThenErrValidation(t *testing.T) {
	meta := domain.DocumentMeta{
		SdocID:            "abc123",
		CreatedAt:         time.Time{},
		UpdatedAt:         time.Now(),
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://example.com",
		UpdateIntervalSec: 86400,
	}

	err := meta.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Contains(t, err.Error(), "CreatedAt")
}

func TestDocumentMeta_WhenZeroUpdatedAt_ThenErrValidation(t *testing.T) {
	meta := domain.DocumentMeta{
		SdocID:            "abc123",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Time{},
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://example.com",
		UpdateIntervalSec: 86400,
	}

	err := meta.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Contains(t, err.Error(), "UpdatedAt")
}

func TestDocumentMeta_WhenBothZeroTimestamps_ThenErrValidation(t *testing.T) {
	meta := domain.DocumentMeta{
		SdocID:            "abc123",
		CreatedAt:         time.Time{},
		UpdatedAt:         time.Time{},
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://example.com",
		UpdateIntervalSec: 86400,
	}

	err := meta.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestDocumentMeta_WhenCreatedAtEqualsUpdatedAt_ThenValid(t *testing.T) {
	now := time.Now()
	meta := domain.DocumentMeta{
		SdocID:            "abc123",
		CreatedAt:         now,
		UpdatedAt:         now,
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://example.com",
		UpdateIntervalSec: 86400,
	}

	err := meta.Validate()

	assert.NoError(t, err)
}

func TestDocumentMeta_WhenZeroOrNegativeUpdateIntervalSec_ThenErrValidation(t *testing.T) {
	now := time.Now()
	for _, interval := range []int{0, -1} {
		meta := domain.DocumentMeta{
			SdocID:            "abc123",
			CreatedAt:         now,
			UpdatedAt:         now,
			SourceID:          "src1",
			Type:              domain.DocumentTypeSearchPage,
			ExternalURL:       "https://example.com",
			UpdateIntervalSec: interval,
		}
		err := meta.Validate()
		assert.ErrorIs(t, err, domain.ErrValidation)
		assert.Contains(t, err.Error(), "UpdateIntervalSec")
	}
}
