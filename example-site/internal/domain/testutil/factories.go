package testutil

import (
	"time"

	"github.com/go-faker/faker/v4"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
)

func RegionFactory() string {
	return faker.Word()
}

func AdvertIdentityFactory(region string) domain.AdvertIdentity {
	if region == "" {
		region = RegionFactory()
	}
	advID := domain.AdvertIdentity{
		ID:     faker.UUIDHyphenated(),
		Region: region,
	}
	return advID
}

func AdvertFactory(region string) domain.Advert {
	ints, _ := faker.RandomInt(1, 1000)
	adv := domain.Advert{
		AdvertIdentity: AdvertIdentityFactory(region),
		Title:          faker.Sentence(),
		Description:    faker.Paragraph(),
		Price:          ints[0],
		PubDate:        time.Unix(faker.RandomUnixTime(), 0),
	}
	return adv
}
