package adverts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/application/adverts/gen"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
	domaintest "github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain/testutil"
)

func newHandlerWithMockRepo(t *testing.T) (*Handler, *domaintest.MockAdvertsRepository) {
	t.Helper()
	repo := domaintest.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(repo)
	require.NoError(t, err)
	return NewHandler(zap.NewNop(), svc), repo
}

func TestHandler_NewError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
		wantDetail bool
	}{
		{
			name:       "nil returns nil",
			err:        nil,
			wantStatus: 0,
			wantMsg:    "",
			wantDetail: false,
		},
		{
			name:       "validation error returns 400 with detail",
			err:        fmt.Errorf("%w: bad input", domain.ErrValidation),
			wantStatus: http.StatusBadRequest,
			wantMsg:    "validation error",
			wantDetail: true,
		},
		{
			name:       "not found returns 404 without detail",
			err:        fmt.Errorf("%w: missing", domain.ErrNotFound),
			wantStatus: http.StatusNotFound,
			wantMsg:    "not found",
			wantDetail: false,
		},
		{
			name:       "unknown error returns 500 without detail",
			err:        errors.New("db connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "internal server error",
			wantDetail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Handler{logger: zap.NewNop()}
			got := h.NewError(context.Background(), tt.err)

			if tt.err == nil {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			assert.Equal(t, tt.wantStatus, got.StatusCode)
			assert.Equal(t, tt.wantMsg, got.Response.Error)
			assert.Equal(t, tt.wantDetail, got.Response.Detail.IsSet())
		})
	}
}

func TestHandler_SearchAdverts_WhenMorePages_ThenNextPageSet(t *testing.T) {
	h, repo := newHandlerWithMockRepo(t)
	region := "us"
	adv := domaintest.AdvertFactory(region)

	repo.EXPECT().SearchAdverts(
		mock.Anything,
		domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 1, OlderFirst: false},
	).Return(domain.AdvertSearchResult{
		AdvertSearchParams: domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 1},
		Adverts:            []domain.Advert{adv},
		AdvertsTotal:       25,
	}, nil)

	resp, err := h.SearchAdverts(context.Background(), gen.SearchAdvertsParams{
		Region: region,
		Size:   gen.NewOptInt(10),
		Page:   gen.NewOptInt(1),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.True(t, resp.NextPage.IsSet())
	v, ok := resp.NextPage.Get()
	require.True(t, ok)
	assert.Equal(t, "/adverts/us/search?page=2&size=10", v)
	assert.Equal(t, adv.ID, resp.Adverts[0].ID)
}

func TestHandler_SearchAdverts_WhenLastPage_ThenNextPageUnset(t *testing.T) {
	h, repo := newHandlerWithMockRepo(t)
	region := "us"
	adv := domaintest.AdvertFactory(region)

	repo.EXPECT().SearchAdverts(
		mock.Anything,
		domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 3, OlderFirst: false},
	).Return(domain.AdvertSearchResult{
		AdvertSearchParams: domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 3},
		Adverts:            []domain.Advert{adv},
		AdvertsTotal:       25,
	}, nil)

	params := gen.SearchAdvertsParams{Region: region, Page: gen.NewOptInt(3), Size: gen.NewOptInt(10)}
	resp, err := h.SearchAdverts(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.False(t, resp.NextPage.IsSet())
	assert.Equal(t, adv.ID, resp.Adverts[0].ID)
}

func TestHandler_SearchAdverts_WhenSpecialChars_ThenURLsEscaped(t *testing.T) {
	h, repo := newHandlerWithMockRepo(t)
	region := "us west"
	adv := domaintest.AdvertFactory(region)

	repo.EXPECT().SearchAdverts(
		mock.Anything,
		mock.MatchedBy(func(p domain.AdvertSearchParams) bool {
			return p.Region == region && p.OlderFirst == false
		}),
	).Return(domain.AdvertSearchResult{
		AdvertSearchParams: domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 1},
		Adverts:            []domain.Advert{adv},
		AdvertsTotal:       1,
	}, nil)

	resp, err := h.SearchAdverts(context.Background(), gen.SearchAdvertsParams{
		Region: region,
		Size:   gen.NewOptInt(10),
		Page:   gen.NewOptInt(1),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Adverts, 1)

	assert.Contains(t, resp.Adverts[0].URL, "/adverts/us%20west/adverts/")
	assert.Equal(t, adv.ID, resp.Adverts[0].ID)
}

func TestHandler_GetAdvert_WhenFound_ThenReturnsDetail(t *testing.T) {
	h, repo := newHandlerWithMockRepo(t)
	adv := domaintest.AdvertFactory("us")

	repo.EXPECT().GetAdvert(mock.Anything, adv.AdvertIdentity).Return(adv, nil)

	resp, err := h.GetAdvert(context.Background(), gen.GetAdvertParams{Region: adv.Region, ID: adv.ID})
	require.NoError(t, err)
	require.NotNil(t, resp)

	detail, ok := resp.(*gen.AdvertDetail)
	require.True(t, ok)
	assert.Equal(t, adv.Title, detail.Title)
	assert.Equal(t, adv.Description, detail.Description)
	assert.Equal(t, adv.Price, detail.Price)
}

func TestHandler_UpsertAdvert_WhenCreated_ThenReturnsCreated(t *testing.T) {
	h, repo := newHandlerWithMockRepo(t)
	adv := domaintest.AdvertFactory("us")

	repo.EXPECT().UpsertAdvert(mock.Anything, adv).Return(true, nil)

	req := &gen.AdvertDetail{
		Title:       adv.Title,
		Description: adv.Description,
		Price:       adv.Price,
		PubDate:     adv.PubDate,
	}
	resp, err := h.UpsertAdvert(context.Background(), req,
		gen.UpsertAdvertParams{Region: adv.Region, ID: adv.ID})
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, ok := resp.(*gen.UpsertAdvertCreated)
	assert.True(t, ok)
}

func TestHandler_UpsertAdvert_WhenUpdated_ThenReturnsOK(t *testing.T) {
	h, repo := newHandlerWithMockRepo(t)
	adv := domaintest.AdvertFactory("us")

	repo.EXPECT().UpsertAdvert(mock.Anything, adv).Return(false, nil)

	req := &gen.AdvertDetail{
		Title:       adv.Title,
		Description: adv.Description,
		Price:       adv.Price,
		PubDate:     adv.PubDate,
	}
	resp, err := h.UpsertAdvert(context.Background(), req,
		gen.UpsertAdvertParams{Region: adv.Region, ID: adv.ID})
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, ok := resp.(*gen.UpsertAdvertOK)
	assert.True(t, ok)
}

func TestHandler_DeleteAdvert_WhenError_ThenPropagates(t *testing.T) {
	h, repo := newHandlerWithMockRepo(t)
	id := domaintest.AdvertIdentityFactory("us")

	repo.EXPECT().DeleteAdvert(mock.Anything, id).Return(errors.New("boom"))

	err := h.DeleteAdvert(context.Background(), gen.DeleteAdvertParams{Region: id.Region, ID: id.ID})
	assert.EqualError(t, err, "boom")
}
