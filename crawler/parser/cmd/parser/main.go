package main

import (
	"context"
	"errors"
	"fmt"
	stdLog "log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vitalyshatskikh/go-lib/closer"
	"github.com/vitalyshatskikh/go-lib/config"
	"github.com/vitalyshatskikh/go-lib/database/postgres"
	"github.com/vitalyshatskikh/go-lib/observability"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/application"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/infrastructure/repositories"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/infrastructure/workers"
)

func main() {
	cfg, err := config.LoadInto(&application.Config{})
	if err != nil {
		stdLog.Fatalf("failed to load config: %v", err)
	}

	logger, err := observability.InitLogger(&cfg.Config)
	if err != nil {
		stdLog.Fatalf("failed to initialize logger: %v", err)
	}

	err = run(cfg, logger)
	if err != nil {
		logger.Error("service exited with errors", zap.Error(err))
		_ = logger.Sync()
		os.Exit(1)
	}
	_ = logger.Sync()
}

func run(cnf *application.Config, logger *zap.Logger) error {
	logger.Info("starting service...")

	// ---- Setup infra ----

	c := closer.New(5 * time.Second) // 5 sec to shutdown
	defer func() {
		logger.Info("shutting down service...")
		if err := c.Close(); err != nil {
			logger.Error("service stopped with errors", zap.Error(err))
		}
		logger.Info("service stopped")
	}()

	dbPool, err := postgres.NewPGXPool(cnf.Postgres, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize db pool: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		dbPool.Close()
		return nil
	})

	// ---- Setup services ----

	slogLogger := slog.New(zapslog.NewHandler(logger.Core()))
	opts := client.Options{
		HostPort:  cnf.TemporalHost,
		Namespace: cnf.TemporalNamespace,
		Logger:    log.NewStructuredLogger(slogLogger),
	}

	temporalClient, err := client.Dial(opts)
	if err != nil {
		return fmt.Errorf("cannot dial temporal: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		temporalClient.Close()
		return nil
	})

	configRepo := repositories.NewPGConfigRepo(dbPool)
	advertsRepo := repositories.NewPGAdvertsRepo(dbPool)

	parsingSvc, err := domain.NewParsingService(configRepo)
	if err != nil {
		return fmt.Errorf("cannot initialize parsing service: %w", err)
	}

	parsingW, err := workers.NewParsingWorker(temporalClient, parsingSvc, advertsRepo)
	if err != nil {
		return fmt.Errorf("cannot create ParsingWorker: %w", err)
	}

	// ---- Run ----

	quitCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err = parsingW.Run(quitCtx)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("ParsingWorker failed: %w", err)
	}

	return nil
}
