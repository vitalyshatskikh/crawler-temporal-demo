package cleaner

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
)

type AdvertsCleaner struct {
	cfg           *CleanConfig
	logger        *zap.Logger
	advertService *domain.AdvertsCRUDService
}

func New(cfg *CleanConfig, logger *zap.Logger, service *domain.AdvertsCRUDService) *AdvertsCleaner {
	return &AdvertsCleaner{
		cfg:           cfg,
		logger:        logger,
		advertService: service,
	}
}

func (c *AdvertsCleaner) Run(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.CleanupInterval)
	defer ticker.Stop()

	c.logger.Info(
		"cleanup started",
		zap.Duration("interval", c.cfg.CleanupInterval),
		zap.Duration("olderThan", c.cfg.CleanupDuration),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("cleanup stopped")
			return
		case <-ticker.C:
			c.logger.Info(
				"running cleanup for adverts",
				zap.Time("older_than", time.Now().Add(-c.cfg.CleanupDuration)),
			)
			err := c.advertService.CleanupDeletedAdverts(ctx, c.cfg.CleanupDuration)
			if err != nil {
				c.logger.Error("failed to cleanup deleted adverts", zap.Error(err))
			}
		}
	}
}
