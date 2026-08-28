package activities_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/application/activities"
	apptestutil "github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/application/testutil"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain/testutil"
)

func TestNewParser_WhenNilSvc_ThenErrValidation(t *testing.T) {
	mockRepo := apptestutil.NewMockAdvertsRepository(t)

	parser, err := activities.NewParser(nil, mockRepo)

	assert.Nil(t, parser)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestNewParser_WhenNilRepo_ThenErrValidation(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)

	parser, err := activities.NewParser(svc, nil)

	assert.Nil(t, parser)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestNewParser_WhenValid_ThenNonNilParser(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewParsingService(mockConfRepo)
	assert.NoError(t, err)

	parser, err := activities.NewParser(svc, mockRepo)

	assert.NoError(t, err)
	assert.NotNil(t, parser)
}

func TestParser_ParseSearchPage_WhenInvalidMetaType_ThenErrValidation(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeSurfedAdvert,
		ExternalURL:       "https://example.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	metas, err := parser.ParseSearchPage(context.Background(), meta)

	assert.Nil(t, metas)
	assert.ErrorIs(t, err, domain.ErrValidation)
	var appErr *temporal.ApplicationError
	assert.ErrorAs(t, err, &appErr)
	assert.True(t, appErr.NonRetryable())
}

func TestParser_ParseSearchPage_WhenRepoGetFails_ThenWrappedError(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.Document{}, errors.New("repo get failed"))

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://search.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	metas, err := parser.ParseSearchPage(context.Background(), meta)

	assert.Nil(t, metas)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get document")
	assert.Contains(t, err.Error(), "repo get failed")
}

func TestParser_ParseSearchPage_WhenServiceFails_ThenWrappedError(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockConfRepo.EXPECT().GetConfig(
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeSearchPage,
		Params:       []domain.ParsingParam{},
	}, nil)

	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "sdoc123",
			SourceID:          "src1",
			Type:              domain.DocumentTypeSearchPage,
			ExternalURL:       "https://search.com",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			UpdateIntervalSec: 86400,
		},
		Body: []byte(`{"urls": []}`),
	}, nil)

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://search.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	metas, err := parser.ParseSearchPage(context.Background(), meta)

	assert.Nil(t, metas)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse search page")
	assert.ErrorIs(t, err, domain.ErrValidation)
	var appErr *temporal.ApplicationError
	assert.ErrorAs(t, err, &appErr)
	assert.True(t, appErr.NonRetryable())
}

func TestParser_ParseSearchPage_WhenValid_ThenSavesAllAndReturnsMetas(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockConfRepo.EXPECT().GetConfig(
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls",
		Params:              []domain.ParsingParam{},
	}, nil)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "parent123",
			SourceID:          "src1",
			Type:              domain.DocumentTypeSearchPage,
			ExternalURL:       "https://search.com",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			UpdateIntervalSec: 86400,
		},
		Body: []byte(`{"urls": ["https://a.com", "https://b.com"]}`),
	}
	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(doc, nil)

	mockRepo.EXPECT().SaveDocument(
		mock.Anything,
		mock.Anything,
	).Return(nil).Twice()

	meta := domain.DocumentMeta{
		SdocID:            "parent123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://search.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	metas, err := parser.ParseSearchPage(context.Background(), meta)

	assert.NoError(t, err)
	assert.NotNil(t, metas)
	assert.Len(t, metas, 2)
	mockRepo.AssertNumberOfCalls(t, "SaveDocument", 2)
}

func TestParser_ParseSearchPage_WhenEmptyURLs_ThenReturnsNonNilEmptySliceAndNoSaves(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockConfRepo.EXPECT().GetConfig(
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls",
		Params:              []domain.ParsingParam{},
	}, nil)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "parent123",
			SourceID:          "src1",
			Type:              domain.DocumentTypeSearchPage,
			ExternalURL:       "https://search.com",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			UpdateIntervalSec: 86400,
		},
		Body: []byte(`{"urls": []}`),
	}
	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(doc, nil)

	meta := domain.DocumentMeta{
		SdocID:            "parent123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://search.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	metas, err := parser.ParseSearchPage(context.Background(), meta)

	assert.NoError(t, err)
	assert.NotNil(t, metas)
	assert.Empty(t, metas)
	mockRepo.AssertNumberOfCalls(t, "SaveDocument", 0)
}

func TestParser_ParseSearchPage_WhenSaveFails_ThenReturnsErrorAndStops(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockConfRepo.EXPECT().GetConfig(
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls",
		Params:              []domain.ParsingParam{},
	}, nil)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "parent123",
			SourceID:          "src1",
			Type:              domain.DocumentTypeSearchPage,
			ExternalURL:       "https://search.com",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			UpdateIntervalSec: 86400,
		},
		Body: []byte(`{"urls": ["https://a.com", "https://b.com"]}`),
	}
	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(doc, nil)

	mockRepo.EXPECT().SaveDocument(
		mock.Anything,
		mock.Anything,
	).Return(errors.New("save failed")).Once()

	meta := domain.DocumentMeta{
		SdocID:            "parent123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://search.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	metas, err := parser.ParseSearchPage(context.Background(), meta)

	assert.Nil(t, metas)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save document")
	assert.Contains(t, err.Error(), "save failed")
	mockRepo.AssertNumberOfCalls(t, "SaveDocument", 1)
}

func TestParser_ParseAdvertContent_WhenInvalidMetaType_ThenErrValidation(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://example.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	err := parser.ParseAdvertContent(context.Background(), meta)

	assert.ErrorIs(t, err, domain.ErrValidation)
	var appErr *temporal.ApplicationError
	assert.ErrorAs(t, err, &appErr)
	assert.True(t, appErr.NonRetryable())
}

func TestParser_ParseAdvertContent_WhenRepoGetFails_ThenWrappedError(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.Document{}, errors.New("repo get failed"))

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeDownloadedAdvert,
		ExternalURL:       "https://example.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	err := parser.ParseAdvertContent(context.Background(), meta)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get document")
	assert.Contains(t, err.Error(), "repo get failed")
}

func TestParser_ParseAdvertContent_WhenServiceFails_ThenWrappedError(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockConfRepo.EXPECT().GetConfig(
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeDownloadedAdvert,
		Params:       []domain.ParsingParam{},
	}, nil)

	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "sdoc123",
			SourceID:          "src1",
			Type:              domain.DocumentTypeDownloadedAdvert,
			ExternalURL:       "https://example.com",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			UpdateIntervalSec: 86400,
		},
		Body: []byte(`{"url": "https://example.com"}`),
	}, nil)

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeDownloadedAdvert,
		ExternalURL:       "https://example.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	err := parser.ParseAdvertContent(context.Background(), meta)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse advert")
	assert.ErrorIs(t, err, domain.ErrValidation)
	var appErr *temporal.ApplicationError
	assert.ErrorAs(t, err, &appErr)
	assert.True(t, appErr.NonRetryable())
}

func TestParser_ParseAdvertContent_WhenSaveFails_ThenWrappedError(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockConfRepo.EXPECT().GetConfig(
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeDownloadedAdvert,
		Params: []domain.ParsingParam{
			{Name: "url", JMESPath: "url", Default: ""},
		},
	}, nil)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "sdoc123",
			SourceID:          "src1",
			Type:              domain.DocumentTypeDownloadedAdvert,
			ExternalURL:       "https://example.com",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			UpdateIntervalSec: 86400,
		},
		Body: []byte(`{"url": "https://example.com/product/123"}`),
	}
	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(doc, nil)

	mockRepo.EXPECT().SaveDocument(
		mock.Anything,
		mock.Anything,
	).Return(errors.New("save failed"))

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeDownloadedAdvert,
		ExternalURL:       "https://example.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	err := parser.ParseAdvertContent(context.Background(), meta)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save document")
	assert.Contains(t, err.Error(), "save failed")
}

func TestParser_ParseAdvertContent_WhenValid_ThenSavesAndReturnsNil(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockConfRepo.EXPECT().GetConfig(
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeDownloadedAdvert,
		Params: []domain.ParsingParam{
			{Name: "url", JMESPath: "url", Default: ""},
		},
	}, nil)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "sdoc123",
			SourceID:          "src1",
			Type:              domain.DocumentTypeDownloadedAdvert,
			ExternalURL:       "https://example.com",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			UpdateIntervalSec: 86400,
		},
		Body: []byte(`{"url": "https://example.com/product/123"}`),
	}
	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(doc, nil)

	mockRepo.EXPECT().SaveDocument(
		mock.Anything,
		mock.Anything,
	).Return(nil).Once()

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeDownloadedAdvert,
		ExternalURL:       "https://example.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	err := parser.ParseAdvertContent(context.Background(), meta)

	assert.NoError(t, err)
	mockRepo.AssertNumberOfCalls(t, "SaveDocument", 1)
}

func TestParser_ParseSearchPage_WhenRepoGetFails_ThenRetryable(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.Document{}, errors.New("repo get failed"))

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://search.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	_, err := parser.ParseSearchPage(context.Background(), meta)

	assert.Error(t, err)
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		assert.False(t, appErr.NonRetryable())
	}
}

func TestParser_ParseAdvertContent_WhenRepoGetFails_ThenRetryable(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.Document{}, errors.New("repo get failed"))

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeDownloadedAdvert,
		ExternalURL:       "https://example.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	err := parser.ParseAdvertContent(context.Background(), meta)

	assert.Error(t, err)
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		assert.False(t, appErr.NonRetryable())
	}
}

func TestParser_ParseSearchPage_WhenSaveFails_ThenRetryable(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockConfRepo.EXPECT().GetConfig(
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.ParsingConfig{
		SourceID:            "src1",
		DocumentType:        domain.DocumentTypeSearchPage,
		ExternalURLJMESPath: "urls",
		Params:              []domain.ParsingParam{},
	}, nil)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "parent123",
			SourceID:          "src1",
			Type:              domain.DocumentTypeSearchPage,
			ExternalURL:       "https://search.com",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			UpdateIntervalSec: 86400,
		},
		Body: []byte(`{"urls": ["https://a.com"]}`),
	}
	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(doc, nil)

	mockRepo.EXPECT().SaveDocument(
		mock.Anything,
		mock.Anything,
	).Return(errors.New("save failed")).Once()

	meta := domain.DocumentMeta{
		SdocID:            "parent123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeSearchPage,
		ExternalURL:       "https://search.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	_, err := parser.ParseSearchPage(context.Background(), meta)

	assert.Error(t, err)
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		assert.False(t, appErr.NonRetryable())
	}
}

func TestParser_ParseAdvertContent_WhenSaveFails_ThenRetryable(t *testing.T) {
	mockConfRepo := testutil.NewMockConfigRepository(t)
	mockRepo := apptestutil.NewMockAdvertsRepository(t)
	svc, _ := domain.NewParsingService(mockConfRepo)
	parser, _ := activities.NewParser(svc, mockRepo)

	mockConfRepo.EXPECT().GetConfig(
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(domain.ParsingConfig{
		SourceID:     "src1",
		DocumentType: domain.DocumentTypeDownloadedAdvert,
		Params: []domain.ParsingParam{
			{Name: "url", JMESPath: "url", Default: ""},
		},
	}, nil)

	doc := domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            "sdoc123",
			SourceID:          "src1",
			Type:              domain.DocumentTypeDownloadedAdvert,
			ExternalURL:       "https://example.com",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			UpdateIntervalSec: 86400,
		},
		Body: []byte(`{"url": "https://example.com/product/123"}`),
	}
	mockRepo.EXPECT().GetDocument(
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(doc, nil)

	mockRepo.EXPECT().SaveDocument(
		mock.Anything,
		mock.Anything,
	).Return(errors.New("save failed"))

	meta := domain.DocumentMeta{
		SdocID:            "sdoc123",
		SourceID:          "src1",
		Type:              domain.DocumentTypeDownloadedAdvert,
		ExternalURL:       "https://example.com",
		ContentURL:        "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UpdateIntervalSec: 86400,
	}

	err := parser.ParseAdvertContent(context.Background(), meta)

	assert.Error(t, err)
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		assert.False(t, appErr.NonRetryable())
	}
}
