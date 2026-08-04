package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/vitalyshatskikh/go-lib/database/postgres"

	"github.com/vitalyshatskikh/go-lib/config"
	"go.uber.org/zap"

	"github.com/vitalyshatskikh/go-lib/closer"
	"github.com/vitalyshatskikh/go-lib/observability"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/application/cleaner"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/infrastructure/repositories"
)

type siteJobCfg struct {
	Base    *config.Config
	Cleaner *cleaner.Config
}

func main() {
	baseCfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load base config: %v", err)
	}

	logger, err := observability.InitLogger(baseCfg)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	cleanCfg := &cleaner.Config{}
	if err := cleanenv.ReadEnv(cleanCfg); err != nil {
		log.Fatalf("failed to read job config: %v", err)
	}

	jobCfg := &siteJobCfg{
		Base:    baseCfg,
		Cleaner: cleanCfg,
	}

	err = run(jobCfg, logger)
	if err != nil {
		logger.Error("service exited with errors", zap.Error(err))
		_ = logger.Sync()
		os.Exit(1)
	}
	_ = logger.Sync()
}

func run(cfg *siteJobCfg, logger *zap.Logger) error {
	logger.Info("starting sitejob...")

	c := closer.New(5 * time.Second)
	defer func() {
		logger.Info("shutting down sitejob...")
		if err := c.Close(); err != nil {
			logger.Error("sitejob stopped with errors", zap.Error(err))
		}
		logger.Info("sitejob stopped")
	}()

	dbPool, err := postgres.NewPGXPool(cfg.Base.Postgres, logger)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	advertsRepo := repositories.NewPGAdvertsRepo(dbPool)
	advertsService, err := domain.NewAdvertsCRUDService(advertsRepo)
	if err != nil {
		return fmt.Errorf("failed to create adverts service: %w", err)
	}

	wg := &sync.WaitGroup{}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	wg.Go(func() {
		cleaner.New(cfg.Cleaner, logger, advertsService).Run(ctx)
	})

	wg.Wait()

	return nil
}
