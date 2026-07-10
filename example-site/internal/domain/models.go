package domain

import "time"

type AdvertIdentity struct {
	ID     string `validate:"required"`
	Region string `validate:"required"`
}

type Advert struct {
	AdvertIdentity
	Title       string `validate:"required"`
	Description string
	Price       int       `validate:"gte=0"`
	PubDate     time.Time `validate:"required"`
}

type AdvertSearchParams struct {
	Region   string `validate:"required"`
	PageSize int    `validate:"gte=1,lte=100"`
	PageNum  int    `validate:"gte=1"`
}

type AdvertSearchResult struct {
	AdvertSearchParams
	Adverts      []Advert
	AdvertsTotal int
}
