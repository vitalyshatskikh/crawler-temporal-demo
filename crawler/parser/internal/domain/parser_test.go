package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain/testutil"
)

func TestService_New_WhenNilConfRepo_ThenErrValidation(t *testing.T) {
	docRepo := testutil.NewMockAdvertsRepository(t)

	svc, err := domain.NewParsingService(nil, docRepo)

	assert.Nil(t, svc)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_New_WhenNilDocRepo_ThenErrValidation(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)

	svc, err := domain.NewParsingService(confRepo, nil)

	assert.Nil(t, svc)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_New_WhenBothValid_ThenNonNilService(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	docRepo := testutil.NewMockAdvertsRepository(t)

	svc, err := domain.NewParsingService(confRepo, docRepo)

	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestService_ParseSearchPage_WhenInvalidMeta_ThenErrValidation(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	docRepo := testutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(confRepo, docRepo)

	meta := domain.DocumentMeta{
		SdocID:      "",
		SourceID:    "src1",
		Type:        domain.DocumentTypeSearchPage,
		ExternalURL: "https://example.com",
	}

	docs, err := svc.ParseSearchPage(context.Background(), meta)

	assert.Nil(t, docs)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseSearchPage_WhenDocRepoError_ThenWrapsErr(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(confRepo, mockDocRepo)

	meta := testutil.MustDocumentMeta("https://search.com", "parent123", "src1")
	mockDocRepo.On("GetDocument", mock.Anything, meta.SdocID, meta.SourceID, domain.DocumentTypeSearchPage).
		Return(domain.Document{}, errors.New("db error"))

	docs, err := svc.ParseSearchPage(context.Background(), meta)

	assert.Nil(t, docs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestService_ParseSearchPage_WhenConfigMissing_ThenErrNotFound(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{}, domain.ErrNotFound)
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	meta := testutil.MustDocumentMeta("https://search.com", "parent123", "src1")

	docs, err := svc.ParseSearchPage(context.Background(), meta)

	assert.Nil(t, docs)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_ParseSearchPage_WhenNoExternalURLParam_ThenErrValidation(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params:       []domain.ParsingParam{{Name: "title", JMESPath: "title", Default: ""}},
	}, nil)
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	meta := testutil.MustDocumentMeta("https://search.com", "parent123", "src1")

	docs, err := svc.ParseSearchPage(context.Background(), meta)

	assert.Nil(t, docs)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseSearchPage_WhenValidInput_ThenUniqueSdocIDsAndCorrectType(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
		},
	}, nil)
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "parent123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeSearchPage,
			ExternalURL: "https://search.com",
		},
		Body: []byte(`{"urls": ["https://a.com", "https://b.com"]}`),
	}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	meta := domain.DocumentMeta{
		SdocID:      domain.SdocID("parent123"),
		SourceID:    domain.SourceID("src1"),
		Type:        domain.DocumentTypeSearchPage,
		ExternalURL: "https://search.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	docs, err := svc.ParseSearchPage(context.Background(), meta)

	assert.NoError(t, err)
	assert.Len(t, docs, 2)
	assert.Equal(t, domain.DocumentTypeSurfedAdvert, docs[0].Type)
	assert.Equal(t, domain.DocumentTypeSurfedAdvert, docs[1].Type)
	assert.NotEqual(t, docs[0].SdocID, docs[1].SdocID)
}

func TestService_ParseSearchPage_WhenEmptySnippets_ThenEmptySlice(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
		},
	}, nil)
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "parent123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeSearchPage,
			ExternalURL: "https://search.com",
		},
		Body: []byte(`{"urls": []}`),
	}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	meta := domain.DocumentMeta{
		SdocID:      domain.SdocID("parent123"),
		SourceID:    domain.SourceID("src1"),
		Type:        domain.DocumentTypeSearchPage,
		ExternalURL: "https://search.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	docs, err := svc.ParseSearchPage(context.Background(), meta)

	assert.NoError(t, err)
	assert.Empty(t, docs)
}

func TestService_ParseAdvertContent_WhenInvalidMeta_ThenErrValidation(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	docRepo := testutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(confRepo, docRepo)

	meta := domain.DocumentMeta{
		SdocID:      "doc123",
		SourceID:    "src1",
		Type:        domain.DocumentTypeDownloadedAdvert,
		ExternalURL: "",
	}

	doc, err := svc.ParseAdvertContent(context.Background(), meta)

	assert.Equal(t, domain.Document{}, doc)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseAdvertContent_WhenDocRepoError_ThenWrapsErr(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(confRepo, mockDocRepo)

	meta := domain.DocumentMeta{
		SdocID:      "doc123",
		SourceID:    "src1",
		Type:        domain.DocumentTypeDownloadedAdvert,
		ExternalURL: "https://example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockDocRepo.On("GetDocument", mock.Anything, meta.SdocID, meta.SourceID, domain.DocumentTypeDownloadedAdvert).
		Return(domain.Document{}, errors.New("db error"))

	doc, err := svc.ParseAdvertContent(context.Background(), meta)

	assert.Equal(t, domain.Document{}, doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestService_ParseAdvertContent_WhenConfigNotFound_ThenErrNotFound(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{}, domain.ErrNotFound)
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	meta := domain.DocumentMeta{
		SdocID:      "doc123",
		SourceID:    "src1",
		Type:        domain.DocumentTypeDownloadedAdvert,
		ExternalURL: "https://example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	doc, err := svc.ParseAdvertContent(context.Background(), meta)

	assert.Equal(t, domain.Document{}, doc)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_ParseAdvertContent_WhenValid_ThenCorrectTypeAndBody(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeDownloadedAdvert,
		Params:       []domain.ParsingParam{{Name: "url", JMESPath: "url", Default: ""}},
	}, nil)
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "doc123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeDownloadedAdvert,
			ExternalURL: "https://example.com",
		},
		Body: []byte(`{"url": "https://example.com/product/123"}`),
	}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	meta := domain.DocumentMeta{
		SdocID:      "doc123",
		SourceID:    "src1",
		Type:        domain.DocumentTypeDownloadedAdvert,
		ExternalURL: "https://example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	doc, err := svc.ParseAdvertContent(context.Background(), meta)

	assert.NoError(t, err)
	assert.Equal(t, domain.DocumentTypeParsedAdvert, doc.Type)
	assert.Equal(t, meta.SdocID, doc.SdocID)
}

func TestService_ParseSearchPage_WhenWrongMetaType_ThenErrValidation(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	docRepo := testutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(confRepo, docRepo)

	meta := domain.DocumentMeta{
		SdocID:      "doc123",
		SourceID:    "src1",
		Type:        domain.DocumentTypeSurfedAdvert,
		ExternalURL: "https://example.com",
	}

	docs, err := svc.ParseSearchPage(context.Background(), meta)

	assert.Nil(t, docs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DocumentTypeSearchPage")
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseAdvertContent_WhenWrongMetaType_ThenErrValidation(t *testing.T) {
	confRepo := testutil.NewMockConfigRepository(t)
	docRepo := testutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(confRepo, docRepo)

	meta := domain.DocumentMeta{
		SdocID:      "doc123",
		SourceID:    "src1",
		Type:        domain.DocumentTypeSurfedAdvert,
		ExternalURL: "https://example.com",
	}

	doc, err := svc.ParseAdvertContent(context.Background(), meta)

	assert.Equal(t, domain.Document{}, doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DocumentTypeDownloadedAdvert")
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestService_ParseSearchPage_WhenPropertyHasFewerValuesThanURLs_ThenNoPanicAndSkipsProperty(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
			{Name: "title", JMESPath: "titles", Default: ""},
		},
	}, nil)
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "parent123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeSearchPage,
			ExternalURL: "https://search.com",
		},
		Body: []byte(`{"urls": ["https://a.com", "https://b.com", "https://c.com"], "titles": ["Only One Title"]}`),
	}, nil)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	meta := testutil.MustDocumentMeta("https://search.com", "parent123", "src1")

	docs, err := svc.ParseSearchPage(context.Background(), meta)

	assert.NoError(t, err)
	assert.Len(t, docs, 3)
	assert.Equal(t, "https://a.com", docs[0].ExternalURL)
	assert.Equal(t, "https://b.com", docs[1].ExternalURL)
	assert.Equal(t, "https://c.com", docs[2].ExternalURL)
}

func TestService_ParseSearchPage_WhenSecondCall_ThenConfigRepoNotCalledAgain(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
		},
	}, nil).Once()
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "parent123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeSearchPage,
			ExternalURL: "https://search.com",
		},
		Body: []byte(`{"urls": ["https://a.com"]}`),
	}, nil).Times(2)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	meta := testutil.MustDocumentMeta("https://search.com", "parent123", "src1")

	_, err1 := svc.ParseSearchPage(context.Background(), meta)
	assert.NoError(t, err1)
	_, err2 := svc.ParseSearchPage(context.Background(), meta)
	assert.NoError(t, err2)

	mockConfRepo.AssertExpectations(t)
}

func TestService_ParseSearchPage_WhenConcurrentFirstLoad_ThenConfigLoadedOnce(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockConfRepo.EXPECT().GetConfig(mock.Anything, mock.Anything, mock.Anything).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls", Default: ""},
		},
	}, nil).Once()
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:      "parent123",
			SourceID:    "src1",
			Type:        domain.DocumentTypeSearchPage,
			ExternalURL: "https://search.com",
		},
		Body: []byte(`{"urls": ["https://a.com"]}`),
	}, nil).Times(3)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	meta := testutil.MustDocumentMeta("https://search.com", "parent123", "src1")

	errCh := make(chan error, 3)
	for range 3 {
		go func() {
			_, err := svc.ParseSearchPage(context.Background(), meta)
			errCh <- err
		}()
	}

	for range 3 {
		err := <-errCh
		assert.NoError(t, err)
	}

	mockConfRepo.AssertExpectations(t)
}

func TestService_GetJMESParser_WhenCtxCancelledBefore_ThenReturnsCtxErr(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockDocRepo := testutil.NewMockAdvertsRepository(t)
	mockDocRepo.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.Document{}, context.Canceled)
	svc, _ := domain.NewParsingService(mockConfRepo, mockDocRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.ParseSearchPage(ctx, domain.DocumentMeta{
		SdocID:      "doc123",
		SourceID:    "src1",
		Type:        domain.DocumentTypeSearchPage,
		ExternalURL: "https://example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	assert.ErrorIs(t, err, context.Canceled)
}
