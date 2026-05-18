package parsing

import (
	"context"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/adverts"
)

var _ Parser = (*DummyParser)(nil)

type DummyParser struct {
	ParseSearchPageSdocIDs  []adverts.SdocID
	ParseSearchPageError    error
	ParseAdvertContentError error
}

func (p *DummyParser) ParseSearchPage(_ context.Context, _ adverts.DocumentMeta) ([]adverts.SdocID, error) {
	return p.ParseSearchPageSdocIDs, p.ParseSearchPageError
}

func (p *DummyParser) ParseAdvertContent(_ context.Context, _ adverts.DocumentMeta) error {
	return p.ParseAdvertContentError
}
