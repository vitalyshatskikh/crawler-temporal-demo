package application

//go:generate go tool mockery

import (
	"context"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

type AdvertsRepository interface {
	GetDocument(ctx context.Context, sdocID domain.SdocID, sourceID domain.SourceID, docType domain.DocumentType) (domain.Document, error)
	SaveDocument(ctx context.Context, doc domain.Document) error
}
