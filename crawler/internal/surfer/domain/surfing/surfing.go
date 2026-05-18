package surfing

import (
	"context"
)

const URLTemplatePageParam = "page"

type TemplateContext struct {
	Values  map[string]string
	Comment string
}

type Params struct {
	ID                int
	Name              string `validate:"required"`
	URLTemplate       string `validate:"required"`
	URLTemplateParams []TemplateContext
	MaxPages          int `validate:"gt=0"`
}

type Repository interface {
	GetSurfConfig(ctx context.Context, name string) (Params, error)
}
