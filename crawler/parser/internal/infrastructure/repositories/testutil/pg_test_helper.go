//go:build integration

package testutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/vitalyshatskikh/go-lib/config"
	"github.com/vitalyshatskikh/go-lib/database/postgres"
)

var TestPool *pgxpool.Pool

var adminCfg config.PostgresConfig

func Setup() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	adminPool, err := postgres.NewPGXPool(cfg.Postgres, zap.NewNop())
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer adminPool.Close()

	testDBName := fmt.Sprintf("parser-%d", os.Getpid())
	_, err = adminPool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE \"%s\"", testDBName))
	if err != nil {
		return fmt.Errorf("failed to create test database %s: %w", testDBName, err)
	}

	adminCfg = cfg.Postgres

	testCfg := cfg.Postgres
	testCfg.Database = testDBName
	TestPool, err = postgres.NewPGXPool(testCfg, zap.NewNop())
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	schema, err := readSchema()
	if err != nil {
		TestPool.Close()
		TestPool = nil
		return fmt.Errorf("failed to read schema: %w", err)
	}

	_, err = TestPool.Exec(context.Background(), schema)
	if err != nil {
		TestPool.Close()
		TestPool = nil
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	return nil
}

func Teardown() {
	if TestPool != nil {
		TestPool.Close()
		TestPool = nil
	}

	dropTestDB(adminCfg)
}

func readSchema() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("failed to get caller")
	}
	schemaPath := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "db", "schema.sql")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func dropTestDB(cfg config.PostgresConfig) {
	adminPool, err := postgres.NewPGXPool(cfg, zap.NewNop())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect for cleanup: %v\n", err)
		return
	}
	defer adminPool.Close()

	_, _ = adminPool.Exec(context.Background(), fmt.Sprintf(`
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = 'parser-%d' AND pid <> pg_backend_pid()
	`, os.Getpid()))

	_, err = adminPool.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s", fmt.Sprintf("parser-%d", os.Getpid())))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to drop test database: %v\n", err)
	}
}
