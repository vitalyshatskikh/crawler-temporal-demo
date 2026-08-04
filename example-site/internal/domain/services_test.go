package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain/testutil"
)

func TestNewAdvertsCRUDService_WhenValidRepo_ThenReturnsService(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)

	svc, err := domain.NewAdvertsCRUDService(mockRepo)

	require.NoError(t, err)
	require.NotNil(t, svc)
}

func TestNewAdvertsCRUDService_WhenNilRepo_ThenReturnsErrValidation(t *testing.T) {
	svc, err := domain.NewAdvertsCRUDService(nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Nil(t, svc)
}

func TestAdvertsCRUDService_GetAdvert_WhenValidID_ThenReturnsAdvert(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	advert := testutil.AdvertFactory("")
	mockRepo.EXPECT().GetAdvert(context.Background(), advert.AdvertIdentity).Return(advert, nil)

	got, err := svc.GetAdvert(context.Background(), advert.AdvertIdentity)

	require.NoError(t, err)
	assert.Equal(t, advert, got)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_GetAdvert_WhenRepoError_ThenReturnsError(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	id := testutil.AdvertIdentityFactory("")
	mockRepo.EXPECT().GetAdvert(context.Background(), id).Return(domain.Advert{}, domain.ErrNotFound)

	got, err := svc.GetAdvert(context.Background(), id)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Equal(t, domain.Advert{}, got)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_GetAdvert_WhenInvalidID_ThenReturnsErrValidation(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	tests := []struct {
		name string
		id   domain.AdvertIdentity
	}{
		{name: "empty ID", id: domain.AdvertIdentity{Region: "test"}},
		{name: "empty region", id: domain.AdvertIdentity{ID: "test"}},
		{name: "both empty", id: domain.AdvertIdentity{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetAdvert(context.Background(), tt.id)

			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrValidation)
			assert.Equal(t, domain.Advert{}, got)
		})
	}
}

func TestAdvertsCRUDService_SearchAdverts_WhenValidParams_ThenReturnsResult(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	region := testutil.RegionFactory()
	params := domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 1}
	adverts := []domain.Advert{testutil.AdvertFactory(region), testutil.AdvertFactory(region)}
	expectedResult := domain.AdvertSearchResult{
		AdvertSearchParams: params,
		Adverts:            adverts,
		AdvertsTotal:       2,
	}
	mockRepo.EXPECT().SearchAdverts(context.Background(), params).Return(expectedResult, nil)

	result, err := svc.SearchAdverts(context.Background(), params)

	require.NoError(t, err)
	assert.Equal(t, params, result.AdvertSearchParams)
	assert.Equal(t, adverts, result.Adverts)
	assert.Equal(t, 2, result.AdvertsTotal)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_SearchAdverts_WhenRepoError_ThenReturnsError(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	params := domain.AdvertSearchParams{Region: testutil.RegionFactory(), PageSize: 10, PageNum: 1}
	mockRepo.EXPECT().SearchAdverts(context.Background(), params).Return(domain.AdvertSearchResult{}, domain.ErrNotFound)

	result, err := svc.SearchAdverts(context.Background(), params)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Equal(t, domain.AdvertSearchResult{}, result)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_SearchAdverts_WhenInvalidParams_ThenReturnsErrValidation(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	tests := []struct {
		name   string
		params domain.AdvertSearchParams
	}{
		{name: "empty region", params: domain.AdvertSearchParams{Region: "", PageSize: 10, PageNum: 1}},
		{name: "page size too small", params: domain.AdvertSearchParams{Region: "test", PageSize: 0, PageNum: 1}},
		{name: "page size too large", params: domain.AdvertSearchParams{Region: "test", PageSize: 200, PageNum: 1}},
		{name: "page num too small", params: domain.AdvertSearchParams{Region: "test", PageSize: 10, PageNum: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.SearchAdverts(context.Background(), tt.params)

			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrValidation)
			assert.Equal(t, domain.AdvertSearchResult{}, result)
		})
	}
}

func TestAdvertsCRUDService_SearchAdverts_WhenNoAdverts_ThenReturnsEmptyResult(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	params := domain.AdvertSearchParams{Region: testutil.RegionFactory(), PageSize: 10, PageNum: 1}
	expectedResult := domain.AdvertSearchResult{
		AdvertSearchParams: params,
		Adverts:            []domain.Advert{},
		AdvertsTotal:       0,
	}
	mockRepo.EXPECT().SearchAdverts(context.Background(), params).Return(expectedResult, nil)

	result, err := svc.SearchAdverts(context.Background(), params)

	require.NoError(t, err)
	assert.Equal(t, params, result.AdvertSearchParams)
	assert.Empty(t, result.Adverts)
	assert.Equal(t, 0, result.AdvertsTotal)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_UpsertAdvert_WhenCreate_ThenReturnsTrue(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	advert := testutil.AdvertFactory("")
	mockRepo.EXPECT().UpsertAdvert(context.Background(), advert).Return(true, nil)

	created, err := svc.UpsertAdvert(context.Background(), advert)

	require.NoError(t, err)
	assert.True(t, created)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_UpsertAdvert_WhenUpdate_ThenReturnsFalse(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	advert := testutil.AdvertFactory("")
	mockRepo.EXPECT().UpsertAdvert(context.Background(), advert).Return(false, nil)

	created, err := svc.UpsertAdvert(context.Background(), advert)

	require.NoError(t, err)
	assert.False(t, created)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_UpsertAdvert_WhenRepoError_ThenReturnsError(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	advert := testutil.AdvertFactory("")
	mockRepo.EXPECT().UpsertAdvert(context.Background(), advert).Return(false, domain.ErrNotFound)

	created, err := svc.UpsertAdvert(context.Background(), advert)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.False(t, created)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_UpsertAdvert_WhenInvalidAdvert_ThenReturnsErrValidation(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	tests := []struct {
		name   string
		advert domain.Advert
	}{
		{name: "empty ID", advert: domain.Advert{AdvertIdentity: domain.AdvertIdentity{Region: "test"}, Title: "title", Price: 10}},
		{name: "empty region", advert: domain.Advert{AdvertIdentity: domain.AdvertIdentity{ID: "test"}, Title: "title", Price: 10}},
		{name: "empty title", advert: domain.Advert{AdvertIdentity: testutil.AdvertIdentityFactory(""), Title: "", Price: 10}},
		{name: "negative price", advert: domain.Advert{AdvertIdentity: testutil.AdvertIdentityFactory(""), Title: "title", Price: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := svc.UpsertAdvert(context.Background(), tt.advert)

			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrValidation)
			assert.False(t, created)
		})
	}
}

func TestAdvertsCRUDService_DeleteAdvert_WhenValidID_ThenReturnsNil(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	id := testutil.AdvertIdentityFactory("")
	mockRepo.EXPECT().DeleteAdvert(context.Background(), id).Return(nil)

	err = svc.DeleteAdvert(context.Background(), id)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_DeleteAdvert_WhenRepoError_ThenReturnsError(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	id := testutil.AdvertIdentityFactory("")
	mockRepo.EXPECT().DeleteAdvert(context.Background(), id).Return(domain.ErrNotFound)

	err = svc.DeleteAdvert(context.Background(), id)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestAdvertsCRUDService_DeleteAdvert_WhenInvalidID_ThenReturnsErrValidation(t *testing.T) {
	mockRepo := testutil.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(mockRepo)
	require.NoError(t, err)

	tests := []struct {
		name string
		id   domain.AdvertIdentity
	}{
		{name: "empty ID", id: domain.AdvertIdentity{Region: "test"}},
		{name: "empty region", id: domain.AdvertIdentity{ID: "test"}},
		{name: "both empty", id: domain.AdvertIdentity{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteAdvert(context.Background(), tt.id)

			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrValidation)
		})
	}
}
