package seller

import (
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
)

func GenerateAdvert(region string) domain.Advert {
	ints, _ := faker.RandomInt(1, 1000)
	return domain.Advert{
		AdvertIdentity: domain.AdvertIdentity{
			ID:     uuid.New().String(),
			Region: region,
		},
		Title:       faker.Sentence(),
		Description: faker.Paragraph(),
		Price:       ints[0],
		PubDate:     time.Now(),
	}
}
