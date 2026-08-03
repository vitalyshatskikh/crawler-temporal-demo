package cleaner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
	domaintest "github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain/testutil"
)

func newCleanerWithMockRepo(t *testing.T, cfg *Config) (*AdvertsCleaner, *domaintest.MockAdvertsRepository) {
	t.Helper()
	repo := domaintest.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(repo)
	require.NoError(t, err)
	return New(cfg, zap.NewNop(), svc), repo
}

func TestAdvertsCleaner_New_WhenValidArgs_ThenReturnsCleaner(t *testing.T) {
	cfg := &Config{
		CleanupInterval: time.Minute,
		CleanupDuration: time.Hour,
	}
	repo := domaintest.NewMockAdvertsRepository(t)
	svc, err := domain.NewAdvertsCRUDService(repo)
	require.NoError(t, err)

	cleaner := New(cfg, zap.NewNop(), svc)

	assert.NotNil(t, cleaner)
}

func TestAdvertsCleaner_Run_WhenTickerFires_ThenCallsCleanupWithDuration(t *testing.T) {
	cfg := &Config{
		CleanupInterval: 10 * time.Millisecond,
		CleanupDuration: time.Hour,
	}
	cleaner, repo := newCleanerWithMockRepo(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo.EXPECT().
		CleanupDeletedAdverts(mock.Anything, time.Hour).
		Return(nil).
		Maybe()

	done := make(chan struct{})
	go func() {
		cleaner.Run(ctx)
		close(done)
	}()

	require.Eventually(t, func() bool {
		return repo.AssertExpectations(t)
	}, time.Second, 10*time.Millisecond)

	cancel()
	<-done
}

func TestAdvertsCleaner_Run_WhenCleanupErrors_ThenContinuesRunning(t *testing.T) {
	cfg := &Config{
		CleanupInterval: 20 * time.Millisecond,
		CleanupDuration: 2 * time.Hour,
	}
	cleaner, repo := newCleanerWithMockRepo(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo.EXPECT().
		CleanupDeletedAdverts(mock.Anything, 2*time.Hour).
		Return(errors.New("db error")).
		Maybe()

	done := make(chan struct{})
	go func() {
		cleaner.Run(ctx)
		close(done)
	}()

	require.Eventually(t, func() bool {
		return repo.AssertExpectations(t)
	}, 2*time.Second, 10*time.Millisecond)

	cancel()
	<-done
}

func TestAdvertsCleaner_Run_WhenContextCancelled_ThenReturns(t *testing.T) {
	cfg := &Config{
		CleanupInterval: time.Hour,
		CleanupDuration: time.Hour,
	}
	cleaner, _ := newCleanerWithMockRepo(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		cleaner.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}
