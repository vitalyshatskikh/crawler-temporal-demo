package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

func TestNewJMESParser_WhenInvalidJMESPath_ThenErrValidation(t *testing.T) {
	tests := []struct {
		name string
		cfg  domain.ParsingConfig
	}{
		{
			name: "nonsense expression",
			cfg: domain.ParsingConfig{
				SourceID:     "src1",
				DocumentType: domain.DocumentTypeSearchPage,
				Params: []domain.ParsingParam{
					{Name: "url", JMESPath: "[invalid", Default: ""},
				},
			},
		},
		{
			name: "empty expression",
			cfg: domain.ParsingConfig{
				SourceID:     "src1",
				DocumentType: domain.DocumentTypeSearchPage,
				Params: []domain.ParsingParam{
					{Name: "url", JMESPath: "", Default: ""},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := domain.NewJMESParser(tt.cfg)
			assert.Nil(t, parser)
			assert.ErrorIs(t, err, domain.ErrValidation)
		})
	}
}

func TestNewJMESParser_WhenEmptyConfig_ThenErrValidation(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params:       []domain.ParsingParam{},
	}

	parser, err := domain.NewJMESParser(cfg)
	assert.Nil(t, parser)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestJMESParser_WhenValidJSONWithSliceResult_ThenCorrectMap(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: "urls", JMESPath: "urls", Default: ""},
		},
	}
	parser, err := domain.NewJMESParser(cfg)
	assert.NoError(t, err)

	result, err := parser.Parse(context.Background(), []byte(`{"urls": ["a","b"]}`))
	assert.NoError(t, err)
	assert.Equal(t, []any{"a", "b"}, result["urls"])
}

func TestJMESParser_WhenNonJSONBody_ThenErrParsingFailed(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: "url", JMESPath: "url", Default: ""},
		},
	}
	parser, err := domain.NewJMESParser(cfg)
	assert.NoError(t, err)

	_, err = parser.Parse(context.Background(), []byte("not json <html>"))
	assert.ErrorIs(t, err, domain.ErrParsingFailed)
}

func TestJMESParser_WhenMissingPath_ThenUsesDefault(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: "url", JMESPath: "items[*].url", Default: "none"},
		},
	}
	parser, err := domain.NewJMESParser(cfg)
	assert.NoError(t, err)

	result, err := parser.Parse(context.Background(), []byte(`{}`))
	assert.NoError(t, err)
	assert.Equal(t, []any{"none"}, result["url"])
}

func TestJMESParser_WhenNilJMESResult_ThenUsesDefault(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: "url", JMESPath: "external_url", Default: "missing"},
		},
	}
	parser, err := domain.NewJMESParser(cfg)
	assert.NoError(t, err)

	result, err := parser.Parse(context.Background(), []byte(`{"external_url": null}`))
	assert.NoError(t, err)
	assert.Equal(t, []any{"missing"}, result["url"])
}

func TestJMESParser_WhenScalarResult_ThenWrappedInSlice(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: "title", JMESPath: "title", Default: ""},
		},
	}
	parser, err := domain.NewJMESParser(cfg)
	assert.NoError(t, err)

	result, err := parser.Parse(context.Background(), []byte(`{"title": "My Title"}`))
	assert.NoError(t, err)
	assert.Equal(t, []any{"My Title"}, result["title"])
}

func TestJMESParser_WhenContextCancelled_ThenCtxErr(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: "url", JMESPath: "url", Default: ""},
		},
	}
	parser, err := domain.NewJMESParser(cfg)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = parser.Parse(ctx, []byte(`{"url": "https://example.com"}`))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestJMESParser_WhenEmptyBody_ThenErrParsingFailed(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: "title", JMESPath: "title", Default: "none"},
		},
	}
	parser, err := domain.NewJMESParser(cfg)
	assert.NoError(t, err)

	_, err = parser.Parse(context.Background(), []byte(""))

	assert.ErrorIs(t, err, domain.ErrParsingFailed)
}

func TestJMESParser_WhenBooleanResult_ThenWrappedInSlice(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: "flag", JMESPath: "active", Default: ""},
		},
	}
	parser, err := domain.NewJMESParser(cfg)
	assert.NoError(t, err)

	result, err := parser.Parse(context.Background(), []byte(`{"active": true}`))

	assert.NoError(t, err)
	assert.Equal(t, []any{true}, result["flag"])
}

func TestJMESParser_WhenNumericResult_ThenWrappedInSlice(t *testing.T) {
	cfg := domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: "price", JMESPath: "price", Default: ""},
		},
	}
	parser, err := domain.NewJMESParser(cfg)
	assert.NoError(t, err)

	result, err := parser.Parse(context.Background(), []byte(`{"price": 42.5}`))

	assert.NoError(t, err)
	assert.Equal(t, []any{42.5}, result["price"])
}
