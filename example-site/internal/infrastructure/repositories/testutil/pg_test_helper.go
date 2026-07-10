//go:build integration

package testutil

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/vitalyshatskikh/go-lib/config"
	"github.com/vitalyshatskikh/go-lib/database/postgres"
)

var TestPool *pgxpool.Pool

var (
	adminCfg   config.PostgresConfig
	testDBName string
)

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

	testDBName = fmt.Sprintf("example-site-%d", os.Getpid())
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

	if err := runMigrations(testCfg); err != nil {
		TestPool.Close()
		TestPool = nil
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func Teardown() {
	if TestPool != nil {
		TestPool.Close()
		TestPool = nil
	}

	if testDBName != "" {
		dropTestDB(adminCfg, testDBName)
	}
}

func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrations")
}

func runMigrations(cfg config.PostgresConfig) error {
	host := cfg.Hosts[0]

	dsn := (&url.URL{
		Scheme: "pgx5",
		User:   url.UserPassword(cfg.User, string(cfg.Password)),
		Host:   host,
		Path:   cfg.Database,
	}).String() + fmt.Sprintf("?sslmode=%s", cfg.SSLMode)

	sourceURL := fmt.Sprintf("file://%s", migrationsDir())
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("cannot create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("cannot run migrations: %w", err)
	}
	return nil
}

func dropTestDB(cfg config.PostgresConfig, dbName string) {
	adminPool, err := postgres.NewPGXPool(cfg, zap.NewNop())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect for cleanup: %v\n", err)
		return
	}
	defer adminPool.Close()

	_, _ = adminPool.Exec(context.Background(), fmt.Sprintf(`
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = '%s' AND pid <> pg_backend_pid()
	`, dbName))

	_, err = adminPool.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to drop test database %s: %v\n", dbName, err)
	}
}
