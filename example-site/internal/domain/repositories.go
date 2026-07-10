//go:generate go tool mockery
package domain

import (
	"context"
)

type AdvertsRepository interface {
	GetAdvert(ctx context.Context, id AdvertIdentity) (Advert, error)
	SearchAdverts(ctx context.Context, params AdvertSearchParams) (AdvertSearchResult, error)
	UpsertAdvert(ctx context.Context, advert Advert) (bool, error)
	DeleteAdvert(ctx context.Context, id AdvertIdentity) error
}
