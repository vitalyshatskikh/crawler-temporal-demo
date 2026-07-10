package domain

import "context"

type AdvertsCRUDService struct {
	repo AdvertsRepository
}

func NewAdvertsCRUDService(repo AdvertsRepository) *AdvertsCRUDService {
	return &AdvertsCRUDService{repo: repo}
}

func (s *AdvertsCRUDService) GetAdvert(ctx context.Context, region, id string) (Advert, error) {
	return s.repo.GetAdvert(ctx, region, id)
}

func (s *AdvertsCRUDService) SearchAdverts(ctx context.Context, params AdvertSearchParams) ([]Advert, error) {
	return s.repo.SearchAdverts(ctx, params)
}

func (s *AdvertsCRUDService) UpsertAdvert(ctx context.Context, advert Advert) (bool, error) {
	return s.repo.UpsertAdvert(ctx, advert)
}

func (s *AdvertsCRUDService) DeleteAdvert(ctx context.Context, region, id string) error {
	return s.repo.DeleteAdvert(ctx, region, id)
}
