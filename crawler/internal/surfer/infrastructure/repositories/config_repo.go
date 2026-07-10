package repositories

import (
	"context"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/surfing"
)

var _ surfing.Repository = (*PGConfigRepository)(nil)

type PGConfigRepository struct{}

func (r *PGConfigRepository) GetSurfConfig(ctx context.Context, name string) (surfing.Params, error) {
	// TODO implement me
	return surfing.Params{
		ID:          1,
		Name:        name,
		URLTemplate: "https://example.com/adverts/{{category}}?page={{page}}",
		URLTemplateParams: []surfing.TemplateContext{
			{Values: map[string]string{"category": "x"}, Comment: "x"},
			{Values: map[string]string{"category": "y"}, Comment: "y"},
			{Values: map[string]string{"category": "z"}, Comment: "z"},
		},
		MaxPages: 5,
	}, nil
}
