package adverts

import "context"

var _ Repository = (*DummyAdvertsRepository)(nil)

type DummyAdvertsRepository struct {
	GetDocumentsMetaBySdocIDResult map[SdocID]DocumentMeta
	GetDocumentsMetaBySdocIDError  error
}

func (d *DummyAdvertsRepository) GetDocumentsMetaBySdocID(
	_ context.Context,
	_ []SdocID,
) (map[SdocID]DocumentMeta, error) {
	return d.GetDocumentsMetaBySdocIDResult, d.GetDocumentsMetaBySdocIDError
}
