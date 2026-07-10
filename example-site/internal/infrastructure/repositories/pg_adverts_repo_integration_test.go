//go:build integration

package repositories_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
	domaintest "github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain/testutil"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/infrastructure/repositories"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/infrastructure/repositories/testutil"
)

func TestMain(m *testing.M) {
	if err := testutil.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "setup failed: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	testutil.Teardown()
	os.Exit(code)
}

func insertAdvert(t *testing.T, pool *pgxpool.Pool, advert domain.Advert) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO adverts (id, region, title, description, price, pub_date, version)
		 VALUES ($1, $2, $3, $4, $5, $6, 0)`,
		advert.ID, advert.Region, advert.Title,
		advert.Description, advert.Price, advert.PubDate,
	)
	require.NoError(t, err)
}

func TestPGAdvertsRepo_GetAdvert_WhenFound_ThenReturnsAdvert(t *testing.T) {
	pool := testutil.TestPool
	repo := repositories.NewPGAdvertsRepo(pool)

	advert := domaintest.AdvertFactory("")
	insertAdvert(t, pool, advert)

	got, err := repo.GetAdvert(context.Background(), advert.AdvertIdentity)

	require.NoError(t, err)
	assert.Equal(t, advert.ID, got.ID)
	assert.Equal(t, advert.Region, got.Region)
	assert.Equal(t, advert.Title, got.Title)
	assert.Equal(t, advert.Description, got.Description)
	assert.Equal(t, advert.Price, got.Price)
	assert.True(t, advert.PubDate.Equal(got.PubDate), "pub_date should match")
}

func TestPGAdvertsRepo_GetAdvert_WhenNotFound_ThenReturnsErrNotFound(t *testing.T) {
	pool := testutil.TestPool
	repo := repositories.NewPGAdvertsRepo(pool)

	id := domaintest.AdvertIdentityFactory("")

	_, err := repo.GetAdvert(context.Background(), id)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestPGAdvertsRepo_SearchAdverts_WhenFoundInRegion_ThenReturnsAdverts(t *testing.T) {
	pool := testutil.TestPool
	repo := repositories.NewPGAdvertsRepo(pool)

	region := domaintest.RegionFactory()
	adverts := []domain.Advert{
		domaintest.AdvertFactory(region),
		domaintest.AdvertFactory(region),
		domaintest.AdvertFactory(region),
	}
	for _, a := range adverts {
		insertAdvert(t, pool, a)
	}

	params := domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 1}
	result, err := repo.SearchAdverts(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Adverts, 3)
	assert.Equal(t, 3, result.AdvertsTotal)
}

func TestPGAdvertsRepo_SearchAdverts_WhenNoMatches_ThenReturnsEmpty(t *testing.T) {
	pool := testutil.TestPool
	repo := repositories.NewPGAdvertsRepo(pool)

	params := domain.AdvertSearchParams{Region: domaintest.RegionFactory(), PageSize: 10, PageNum: 1}
	result, err := repo.SearchAdverts(context.Background(), params)

	require.NoError(t, err)
	assert.Empty(t, result.Adverts)
	assert.Equal(t, 0, result.AdvertsTotal)
}

func TestPGAdvertsRepo_SearchAdverts_WhenPagination_ThenReturnsCorrectPage(t *testing.T) {
	pool := testutil.TestPool
	repo := repositories.NewPGAdvertsRepo(pool)

	region := domaintest.RegionFactory()
	total := 25
	for range total {
		insertAdvert(t, pool, domaintest.AdvertFactory(region))
	}

	result1, err := repo.SearchAdverts(context.Background(), domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 1})
	require.NoError(t, err)
	assert.Len(t, result1.Adverts, 10)
	assert.Equal(t, 25, result1.AdvertsTotal)

	result2, err := repo.SearchAdverts(context.Background(), domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 2})
	require.NoError(t, err)
	assert.Len(t, result2.Adverts, 10)
	assert.Equal(t, 25, result2.AdvertsTotal)

	result3, err := repo.SearchAdverts(context.Background(), domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 3})
	require.NoError(t, err)
	assert.Len(t, result3.Adverts, 5)
	assert.Equal(t, 25, result3.AdvertsTotal)

	result4, err := repo.SearchAdverts(context.Background(), domain.AdvertSearchParams{Region: region, PageSize: 10, PageNum: 4})
	require.NoError(t, err)
	assert.Empty(t, result4.Adverts)
	assert.Equal(t, 0, result4.AdvertsTotal, "COUNT(*) OVER() returns 0 on empty result set")
}

func TestPGAdvertsRepo_UpsertAdvert_WhenNewAdvert_ThenCreatesAndReturnsTrue(t *testing.T) {
	pool := testutil.TestPool
	repo := repositories.NewPGAdvertsRepo(pool)

	advert := domaintest.AdvertFactory("")

	created, err := repo.UpsertAdvert(context.Background(), advert)

	require.NoError(t, err)
	assert.True(t, created)
}

func TestPGAdvertsRepo_UpsertAdvert_WhenExistingAdvert_ThenUpdatesAndReturnsFalse(t *testing.T) {
	pool := testutil.TestPool
	repo := repositories.NewPGAdvertsRepo(pool)

	advert := domaintest.AdvertFactory("")
	_, err := repo.UpsertAdvert(context.Background(), advert)
	require.NoError(t, err)

	created, err := repo.UpsertAdvert(context.Background(), advert)

	require.NoError(t, err)
	assert.False(t, created)
}

func TestPGAdvertsRepo_DeleteAdvert_WhenExisting_ThenDeletes(t *testing.T) {
	pool := testutil.TestPool
	repo := repositories.NewPGAdvertsRepo(pool)

	advert := domaintest.AdvertFactory("")
	insertAdvert(t, pool, advert)

	err := repo.DeleteAdvert(context.Background(), advert.AdvertIdentity)

	require.NoError(t, err)

	_, err = repo.GetAdvert(context.Background(), advert.AdvertIdentity)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestPGAdvertsRepo_DeleteAdvert_WhenNotExisting_ThenNoError(t *testing.T) {
	pool := testutil.TestPool
	repo := repositories.NewPGAdvertsRepo(pool)

	id := domaintest.AdvertIdentityFactory("")

	err := repo.DeleteAdvert(context.Background(), id)

	assert.NoError(t, err)
}
