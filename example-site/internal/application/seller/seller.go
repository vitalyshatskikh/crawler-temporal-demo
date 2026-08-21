//go:generate go tool ogen --config .ogen.yml --target gen --package gen --clean ../../../docs/openapi/openapi.yml
package seller

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/application/seller/gen"
)

type Seller struct {
	cfg        Config
	logger     *zap.Logger
	siteClient *gen.Client
}

func New(baseURL string, sellerCfg Config, logger *zap.Logger) (*Seller, error) {
	client, err := gen.NewClient(baseURL)
	if err != nil {
		return nil, err
	}

	seller := &Seller{
		cfg:        sellerCfg,
		logger:     logger,
		siteClient: client,
	}
	return seller, nil
}

func (s *Seller) Run(ctx context.Context) {
	wg := &sync.WaitGroup{}

	wg.Go(func() {
		s.StartCreator(ctx)
	})
	wg.Go(func() {
		s.StartDeleter(ctx)
	})

	wg.Wait()
}

func (s *Seller) StartCreator(ctx context.Context) {
	interval := time.Second / time.Duration(s.cfg.CreateRPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("creator started", zap.String("region", s.cfg.Region), zap.Int("rps", s.cfg.CreateRPS))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("creator stopped", zap.String("region", s.cfg.Region))
			return
		case <-ticker.C:
			if err := s.TryCreateAdvert(ctx); err != nil {
				s.logger.Error("cannot create advert", zap.String("region", s.cfg.Region), zap.Error(err))
			}
		}
	}
}

func (s *Seller) StartDeleter(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.DeleteInterval)
	defer ticker.Stop()

	s.logger.Info(
		"deleter started",
		zap.String("region", s.cfg.Region),
		zap.Duration("age", s.cfg.DeleteAge),
		zap.Duration("jitter", s.cfg.DeleteJitter),
		zap.Duration("interval", s.cfg.DeleteInterval),
		zap.Int("batchSize", s.cfg.DeleteBatchSize),
	)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("deleter stopped", zap.String("region", s.cfg.Region))
			return
		case <-ticker.C:
			if err := s.TryDeleteAdverts(ctx); err != nil {
				s.logger.Error("cannot delete adverts", zap.String("region", s.cfg.Region), zap.Error(err))
			}
		}
	}
}

func (s *Seller) TryCreateAdvert(ctx context.Context) error {
	advert := GenerateAdvert(s.cfg.Region)
	_, err := s.siteClient.UpsertAdvert(
		ctx,
		&gen.AdvertDetail{
			Title:       advert.Title,
			Description: advert.Description,
			Price:       advert.Price,
			PubDate:     advert.PubDate,
		},
		gen.UpsertAdvertParams{
			Region: advert.Region,
			ID:     advert.ID,
		},
	)
	if err != nil {
		s.logger.Error("failed to upsert advert", zap.String("region", s.cfg.Region), zap.Error(err))
		return err
	}
	return nil
}

func (s *Seller) TryDeleteAdverts(ctx context.Context) error {
	jitterOffset := time.Duration(rand.Int63n(int64(s.cfg.DeleteJitter)))

	pageNum := 1
	totalDeleted := 0

	for {
		result, err := s.siteClient.SearchAdverts(ctx, gen.SearchAdvertsParams{
			Region:     s.cfg.Region,
			Size:       gen.NewOptInt(s.cfg.DeleteBatchSize),
			Page:       gen.NewOptInt(pageNum),
			OlderFirst: gen.NewOptBool(true),
		})
		if err != nil {
			s.logger.Error("failed to search adverts to delete", zap.String("region", s.cfg.Region), zap.Error(err))
			return err
		}

		if len(result.Adverts) == 0 {
			return nil
		}

		for _, adv := range result.Adverts {
			if totalDeleted >= s.cfg.DeleteBatchSize {
				return nil
			}

			advDetail, err := s.siteClient.GetAdvert(ctx, gen.GetAdvertParams{Region: s.cfg.Region, ID: adv.ID})
			if err != nil {
				s.logger.Error(
					"failed to get advert details (to delete)",
					zap.String("region", s.cfg.Region),
					zap.String("id", adv.ID),
					zap.Error(err),
				)
				continue
			}

			advAge := time.Since(advDetail.(*gen.AdvertDetail).PubDate)
			if advAge > s.cfg.DeleteAge+jitterOffset {
				err = s.siteClient.DeleteAdvert(ctx, gen.DeleteAdvertParams{Region: s.cfg.Region, ID: adv.ID})
				if err != nil {
					s.logger.Error(
						"failed to delete advert",
						zap.String("region", s.cfg.Region),
						zap.String("id", adv.ID),
						zap.Error(err),
					)
					continue
				}
			}

			totalDeleted++
		}

		if result.Total <= pageNum*s.cfg.DeleteBatchSize {
			return nil
		}
		pageNum++
	}
}
