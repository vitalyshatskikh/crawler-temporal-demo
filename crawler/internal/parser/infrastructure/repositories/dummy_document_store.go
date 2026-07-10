package repositories

import (
	"context"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/parser/domain/adverts"
	surferadverts "github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/adverts"
)

var _ adverts.DocumentStore = (*DummyDocumentStore)(nil)

type DummyDocumentStore struct {
	FetchDocumentBody []byte
	FetchError        error
}

func (ds *DummyDocumentStore) Fetch(_ context.Context, meta surferadverts.DocumentMeta) ([]byte, error) {
	if ds.FetchError != nil {
		return nil, ds.FetchError
	}
	return ds.FetchDocumentBody, nil
}
