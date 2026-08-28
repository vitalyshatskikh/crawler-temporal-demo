package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

func TestParsingConfigValidate_WhenSearchPageWithExternalURLJMESPath_ThenValid(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls[*]",
		Params:              []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestParsingConfigValidate_WhenSearchPageWithoutExternalURLJMESPath_ThenErrValidation(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params:       []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Contains(t, err.Error(), "ExternalURLJMESPath")
}

func TestParsingConfigValidate_WhenDownloadedAdvertWithEmptyExternalURLJMESPath_ThenValid(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeDownloadedAdvert,
		Params:       []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestParsingConfigValidate_WhenDownloadedAdvertWithContentURLTemplate_ThenValid(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeDownloadedAdvert,
		ExternalURLJMESPath: "",
		ContentURLTemplate:  "https://cdn.example.com{{_external_url}}",
		Params:              []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestParsingConfigValidate_WhenMalformedExternalURLTemplate_ThenErrValidation(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls[*]",
		ExternalURLTemplate: "{{_external_url",
		Params:              []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Contains(t, err.Error(), "ExternalURLTemplate")
}

func TestParsingConfigValidate_WhenMalformedContentURLTemplate_ThenErrValidation(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls[*]",
		ContentURLTemplate:  "{{#if}}",
		Params:              []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Contains(t, err.Error(), "ContentURLTemplate")
}

func TestParsingConfigValidate_WhenEmptyExternalURLTemplate_ThenValid(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls[*]",
		ExternalURLTemplate: "",
		Params:              []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestParsingConfigValidate_WhenEmptyContentURLTemplate_ThenValid(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls[*]",
		ContentURLTemplate:  "",
		Params:              []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestParsingConfigValidate_WhenParamNameStartsWithUnderscore_ThenErrValidation(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls[*]",
		Params: []domain.ParsingParam{
			{Name: "_external_url", JMESPath: "url", Default: ""},
		},
	}

	err := cfg.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Contains(t, err.Error(), "_external_url")
}

func TestParsingConfigValidate_WhenParamNameHasUnderscoreInMiddle_ThenValid(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls[*]",
		Params: []domain.ParsingParam{
			{Name: "my_external_url", JMESPath: "url", Default: ""},
		},
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestParsingConfigValidate_WhenValidSearchPageWithParams_ThenValid(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls[*]",
		ExternalURLTemplate: "{{_external_url}}?ref=search",
		ContentURLTemplate:  "",
		Params: []domain.ParsingParam{
			{Name: "title", JMESPath: "titles", Default: ""},
			{Name: "price", JMESPath: "prices", Default: "0"},
		},
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestParsingConfigValidate_WhenEmptySourceID_ThenErrValidation(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "",
		DocumentType: domain.DocumentTypeSearchPage,
		Params:       []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Contains(t, err.Error(), "SourceID")
}

func TestParsingConfigValidate_WhenEmptyDocumentType_ThenErrValidation(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: "",
		Params:       []domain.ParsingParam{},
	}

	err := cfg.Validate()

	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Contains(t, err.Error(), "DocumentType")
}
