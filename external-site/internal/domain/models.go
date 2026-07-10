package domain

import "time"

type AdvertIdentity struct {
	ID     string
	Region string
}

type Advert struct {
	AdvertIdentity
	Title       string
	Description string
	Price       int
	PubDate     time.Time
}

type AdvertSearchParams struct {
	Region   string
	PageSize int
	PageNum  int
}
