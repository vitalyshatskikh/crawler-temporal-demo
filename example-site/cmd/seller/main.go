package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vitalyshatskikh/go-lib/config"
	"go.uber.org/zap"

	"github.com/vitalyshatskikh/go-lib/observability"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/application/seller"
)

type siteJobCfg struct {
	config.Config
	SiteBaseURL string        `env:"BASE_URL"`
	Seller      seller.Config `env-prefix:"SELLER_"`
}

func main() {
	cfg, err := config.LoadInto(&siteJobCfg{})
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := observability.InitLogger(&cfg.Config)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	err = run(cfg, logger)
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
