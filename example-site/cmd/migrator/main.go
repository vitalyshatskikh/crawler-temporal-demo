package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/vitalyshatskikh/go-lib/config"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := run(cfg); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func run(cfg *config.Config) error {
	iofsDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	dbURL, err := url.Parse(cfg.Postgres.ConnString())
	if err != nil {
		return fmt.Errorf("failed to parse postgres connection string: %w", err)
	}
	dbURL.Scheme = "pgx5"

	m, err := migrate.NewWithSourceInstance("iofs", iofsDriver, dbURL.String())
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	version, _, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}
	fmt.Printf("migrations applied successfully, version: %d\n", version)

	return nil
}
