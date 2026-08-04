package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/vitalyshatskikh/go-lib/config"
	"go.uber.org/zap"

	"github.com/vitalyshatskikh/go-lib/observability"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/application/seller"
)

type siteJobCfg struct {
	Base        *config.Config
	SiteBaseURL string        `env:"BASE_URL"`
	Seller      seller.Config `env-prefix:"SELLER_"`
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

	var jobCfg siteJobCfg
	if err := cleanenv.ReadEnv(&jobCfg); err != nil {
		log.Fatalf("failed to read job config: %v", err)
	}

	jobCfg.Base = baseCfg

	err = run(&jobCfg, logger)
	if err != nil {
		logger.Error("service exited with errors", zap.Error(err))
		_ = logger.Sync()
		os.Exit(1)
	}
	_ = logger.Sync()
}

func run(cfg *siteJobCfg, logger *zap.Logger) error {
	logger.Info("starting seller job...")

	sel, err := seller.New(cfg.SiteBaseURL, cfg.Seller, logger)
	if err != nil {
		return fmt.Errorf("failed to create seller: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	sel.Run(ctx)

	return nil
}
