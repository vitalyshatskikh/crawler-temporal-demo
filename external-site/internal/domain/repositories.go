//go:generate go tool mockery
package domain

import (
	"context"
)

type AdvertsRepository interface {
	GetAdvert(ctx context.Context, region, id string) (Advert, error)
	SearchAdverts(ctx context.Context, params AdvertSearchParams) ([]Advert, error)
	UpsertAdvert(ctx context.Context, advert Advert) (bool, error)
	DeleteAdvert(ctx context.Context, region, id string) error
}
