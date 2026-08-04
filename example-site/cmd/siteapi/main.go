package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vitalyshatskikh/go-lib/database/postgres"
	"go.uber.org/zap"

	"github.com/vitalyshatskikh/go-lib/closer"
	"github.com/vitalyshatskikh/go-lib/config"
	"github.com/vitalyshatskikh/go-lib/http/restapi"
	"github.com/vitalyshatskikh/go-lib/observability"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/docs"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/application/adverts"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/infrastructure/repositories"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := observability.InitLogger(cfg)
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

func run(cfg *config.Config, logger *zap.Logger) error {
	logger.Info("starting service...")

	// ---- Setup infra ----

	ctx := context.Background()
	c := closer.New(5 * time.Second) // 5 sec to shutdown
	defer func() {
		logger.Info("shutting down service...")
		if err := c.Close(); err != nil {
			logger.Error("service stopped with errors", zap.Error(err))
		}
		logger.Info("service stopped")
	}()

	shutdownTelemetry, err := observability.InitTelemetry(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	c.Add(shutdownTelemetry)

	shutdownMetrics, err := observability.InitMetrics(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}
	c.Add(shutdownMetrics)

	dbPool, err := postgres.NewPGXPool(cfg.Postgres, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize db pool: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		dbPool.Close()
		return nil
	})

	// ---- Setup services ----

	advertsRepo := repositories.NewPGAdvertsRepo(dbPool)
	advertsService, err := domain.NewAdvertsCRUDService(advertsRepo)
	if err != nil {
		return fmt.Errorf("failed to initialize adverts service: %w", err)
	}

	advertsRouter, err := adverts.NewRouter(logger, advertsService)
	if err != nil {
		return fmt.Errorf("failed to create adverts router: %w", err)
	}

	srv, err := restapi.New(
		cfg,
		restapi.WithLogger(logger),
		restapi.WithOpenAPI(bytes.NewReader(docs.OpenapiYML)),
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	c.Add(srv.Shutdown)

	err = srv.Mount(
		restapi.SubRoute{Prefix: "/", Handler: http.RedirectHandler("/docs", http.StatusFound)},
		restapi.SubRoute{Prefix: "/adverts", Handler: advertsRouter},
	)
	if err != nil {
		return fmt.Errorf("failed to mount routes: %w", err)
	}

	// ---- Run ----

	stop := make(chan error, 1)
	go func() {
		logger.Info("starting api server")
		stop <- srv.Run()
		close(stop)
	}()

	// ---- Stop ----

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case err = <-stop:
	}

	return err
}
