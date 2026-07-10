package adverts

import (
	"context"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/adverts"
)

// DocumentStore fetches raw document body from external storage using DocumentMeta
type DocumentStore interface {
	Fetch(ctx context.Context, meta adverts.DocumentMeta) ([]byte, error)
}
