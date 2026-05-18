package surfing

import "context"

var _ Repository = (*DummyConfigRepository)(nil)

type DummyConfigRepository struct {
	GetSurfConfigResult Params
	GetSurfConfigError  error
}

func (d *DummyConfigRepository) GetSurfConfig(_ context.Context, _ string) (Params, error) {
	return d.GetSurfConfigResult, d.GetSurfConfigError
}
