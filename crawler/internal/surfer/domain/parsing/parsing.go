package parsing

import (
	"context"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/adverts"
)

// Parser is a tool to extract advert properties from downloaded document
type Parser interface {
	ParseSearchPage(ctx context.Context, docMeta adverts.DocumentMeta) ([]adverts.SdocID, error)
	ParseAdvertContent(ctx context.Context, docMeta adverts.DocumentMeta) error
}
