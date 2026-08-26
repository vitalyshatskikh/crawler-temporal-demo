package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain/testutil"
)

func TestNewParsingServiceWhenNilConfRepo_ThenErrValidation(t *testing.T) {
	svc, err := domain.NewParsingService(nil)

	assert.Nil(t, svc)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestNewParsingServiceWhenValidConfRepo_ThenNonNilService(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)

	svc, err := domain.NewParsingService(confRepo)

	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestService_ParseSearchPage_WhenInvalidDoc_ThenErrValidation(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	svc, _ := domain.NewParsingService(confRepo)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "",
			SourceID:    "src1",
			Type:        domain.DocumentTypeSearchPage,
			ExternalURL: "https://example.com",
		},
	}

	docs, err := svc.ParseSearchPage(context.Background(), doc)

	assert.Nil(t, docs)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseSearchPage_WhenConfigMissing_ThenErrNotFound(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{}, domain.ErrNotFound)
	svc, _ := domain.NewParsingService(mockConfRepo)

	doc := domain.Document{
		DocumentMeta: testutil.MustDocumentMeta("https://search.com", "parent123", "src1"),
		Body:         []byte(`{"urls": ["https://a.com"]}`),
	}

	docs, err := svc.ParseSearchPage(context.Background(), doc)

	assert.Nil(t, docs)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_ParseSearchPage_WhenNoExternalURLParam_ThenErrValidation(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params:       []domain.ParsingParam{{Name: "title", JMESPath: "title", Default: ""}},
	}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo)

	doc := domain.Document{
		DocumentMeta: testutil.MustDocumentMeta("https://search.com", "parent123", "src1"),
		Body:         []byte(`{}`),
	}

	docs, err := svc.ParseSearchPage(context.Background(), doc)

	assert.Nil(t, docs)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseSearchPage_WhenValidInput_ThenUniqueSdocIDsAndCorrectType(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
		},
	}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      domain.SdocID("parent123"),
			SourceID:    domain.SourceID("src1"),
			Type:        domain.DocumentTypeSearchPage,
			ExternalURL: "https://search.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Body: []byte(`{"urls": ["https://a.com", "https://b.com"]}`),
	}

	docs, err := svc.ParseSearchPage(context.Background(), doc)

	assert.NoError(t, err)
	assert.Len(t, docs, 2)
	assert.Equal(t, domain.DocumentTypeSurfedAdvert, docs[0].Type)
	assert.Equal(t, domain.DocumentTypeSurfedAdvert, docs[1].Type)
	assert.NotEqual(t, docs[0].SdocID, docs[1].SdocID)
}

func TestService_ParseSearchPage_WhenEmptySnippets_ThenEmptySlice(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
		},
	}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      domain.SdocID("parent123"),
			SourceID:    domain.SourceID("src1"),
			Type:        domain.DocumentTypeSearchPage,
			ExternalURL: "https://search.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Body: []byte(`{"urls": []}`),
	}

	docs, err := svc.ParseSearchPage(context.Background(), doc)

	assert.NoError(t, err)
	assert.Empty(t, docs)
}

func TestService_ParseAdvertContent_WhenInvalidDoc_ThenErrValidation(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	svc, _ := domain.NewParsingService(confRepo)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "doc123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeDownloadedAdvert,
			ExternalURL: "",
		},
	}

	result, err := svc.ParseAdvertContent(context.Background(), doc)

	assert.Equal(t, domain.Document{}, result)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseAdvertContent_WhenConfigNotFound_ThenErrNotFound(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{}, domain.ErrNotFound)
	svc, _ := domain.NewParsingService(mockConfRepo)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "doc123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeDownloadedAdvert,
			ExternalURL: "https://example.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Body: []byte(`{"url": "https://example.com/product/123"}`),
	}

	result, err := svc.ParseAdvertContent(context.Background(), doc)

	assert.Equal(t, domain.Document{}, result)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_ParseAdvertContent_WhenValid_ThenCorrectTypeAndBody(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeDownloadedAdvert,
		Params:       []domain.ParsingParam{{Name: "url", JMESPath: "url", Default: ""}},
	}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "doc123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeDownloadedAdvert,
			ExternalURL: "https://example.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Body: []byte(`{"url": "https://example.com/product/123"}`),
	}

	result, err := svc.ParseAdvertContent(context.Background(), doc)

	assert.NoError(t, err)
	assert.Equal(t, domain.DocumentTypeParsedAdvert, result.Type)
	assert.Equal(t, doc.SdocID, result.SdocID)
}

func TestService_ParseSearchPage_WhenWrongDocType_ThenErrValidation(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	svc, _ := domain.NewParsingService(confRepo)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "doc123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeSurfedAdvert,
			ExternalURL: "https://example.com",
		},
	}

	docs, err := svc.ParseSearchPage(context.Background(), doc)

	assert.Nil(t, docs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DocumentTypeSearchPage")
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseAdvertContent_WhenWrongDocType_ThenErrValidation(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	svc, _ := domain.NewParsingService(confRepo)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "doc123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeSurfedAdvert,
			ExternalURL: "https://example.com",
		},
	}

	result, err := svc.ParseAdvertContent(context.Background(), doc)

	assert.Equal(t, domain.Document{}, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DocumentTypeDownloadedAdvert")
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseSearchPage_WhenPropertyHasFewerValuesThanURLs_ThenNoPanicAndSkipsProperty(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
			{Name: "title", JMESPath: "titles", Default: ""},
		},
	}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo)

	doc := domain.Document{
		DocumentMeta: testutil.MustDocumentMeta("https://search.com", "parent123", "src1"),
		Body:         []byte(`{"urls": ["https://a.com", "https://b.com", "https://c.com"], "titles": ["Only One Title"]}`),
	}

	docs, err := svc.ParseSearchPage(context.Background(), doc)

	assert.NoError(t, err)
	assert.Len(t, docs, 3)
	assert.Equal(t, "https://a.com", docs[0].ExternalURL)
	assert.Equal(t, "https://b.com", docs[1].ExternalURL)
	assert.Equal(t, "https://c.com", docs[2].ExternalURL)
}

func TestService_ParseSearchPage_WhenSecondCall_ThenConfigRepoNotCalledAgain(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
		},
	}, nil).Once()
	svc, _ := domain.NewParsingService(mockConfRepo)

	doc := domain.Document{
		DocumentMeta: testutil.MustDocumentMeta("https://search.com", "parent123", "src1"),
		Body:         []byte(`{"urls": ["https://a.com"]}`),
	}

	_, err1 := svc.ParseSearchPage(context.Background(), doc)
	assert.NoError(t, err1)
	_, err2 := svc.ParseSearchPage(context.Background(), doc)
	assert.NoError(t, err2)

	mockConfRepo.AssertExpectations(t)
}

func TestService_ParseSearchPage_WhenConcurrentFirstLoad_ThenConfigLoadedOnce(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
		},
	}, nil).Once()
	svc, _ := domain.NewParsingService(mockConfRepo)

	doc := domain.Document{
		DocumentMeta: testutil.MustDocumentMeta("https://search.com", "parent123", "src1"),
		Body:         []byte(`{"urls": ["https://a.com"]}`),
	}

	errCh := make(chan error, 3)
	for range 3 {
		go func() {
			_, err := svc.ParseSearchPage(context.Background(), doc)
			errCh <- err
		}()
	}

	for range 3 {
		err := <-errCh
		assert.NoError(t, err)
	}

	mockConfRepo.AssertExpectations(t)
}

func TestService_ParseSearchPage_WhenCtxCancelled_ThenReturnsCtxErr(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "doc123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeSearchPage,
			ExternalURL: "https://example.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Body: []byte(`{"urls": ["https://a.com"]}`),
	}

	_, err := svc.ParseSearchPage(ctx, doc)

	assert.ErrorIs(t, err, context.Canceled)
}
