//go:generate go tool mockery
package domain

import (
	"context"
)

type ConfigRepository interface {
	GetConfig(ctx context.Context, sourceID SourceID, docType DocumentType) (ParsingConfig, error)
}

type AdvertsRepository interface {
	GetDocument(ctx context.Context, sdocID SdocID, sourceID SourceID, docType DocumentType) (Document, error)
	SaveDocument(ctx context.Context, doc Document) error
}
