package domain

import (
	"context"
	"fmt"
)

type AdvertsCRUDService struct {
	repo AdvertsRepository
}

func NewAdvertsCRUDService(repo AdvertsRepository) (*AdvertsCRUDService, error) {
	if repo == nil {
		return nil, fmt.Errorf("%w: repo cannot be nil", ErrValidation)
	}
	return &AdvertsCRUDService{repo: repo}, nil
}

func (s *AdvertsCRUDService) GetAdvert(ctx context.Context, id AdvertIdentity) (Advert, error) {
	if err := ValidateStruct(id); err != nil {
		return Advert{}, err
	}
	return s.repo.GetAdvert(ctx, id)
}

func (s *AdvertsCRUDService) SearchAdverts(ctx context.Context, params AdvertSearchParams) (AdvertSearchResult, error) {
	if err := ValidateStruct(params); err != nil {
		return AdvertSearchResult{}, err
	}
	return s.repo.SearchAdverts(ctx, params)
}

func (s *AdvertsCRUDService) UpsertAdvert(ctx context.Context, advert Advert) (bool, error) {
	if err := ValidateStruct(advert); err != nil {
		return false, err
	}
	return s.repo.UpsertAdvert(ctx, advert)
}

func (s *AdvertsCRUDService) DeleteAdvert(ctx context.Context, id AdvertIdentity) error {
	if err := ValidateStruct(id); err != nil {
		return err
	}
	return s.repo.DeleteAdvert(ctx, id)
}
