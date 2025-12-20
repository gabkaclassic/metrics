// Package storage provides data storage implementations for metrics.
//
// It supports two storage backends:
//   - MemStorage: In-memory storage for development and testing
//   - DBStorage: PostgreSQL database storage for production use
//
// The package handles storage initialization, connection management,
// and database migrations.
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gabkaclassic/metrics/internal/config"
	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// MemStorage provides in-memory storage for metrics.
// Suitable for development, testing, or single-instance deployments.
type MemStorage struct {
	// Metrics stores metrics in a map with metric ID as key.
	Metrics map[string]models.Metrics
}

// NewMemStorage creates and initializes a new in-memory storage.
// Returns a ready-to-use MemStorage with an empty metrics map.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]models.Metrics),
	}
}

// NewDBStorage creates and initializes a PostgreSQL database connection.
//
// cfg: Database configuration containing DSN, driver, and migration settings.
//
// Returns:
//   - *sql.DB: Established database connection
//   - error: Connection or migration failure details
//
// The function:
//  1. Validates configuration (DSN is required)
//  2. Opens and verifies database connection
//  3. Runs database migrations if configured
//  4. Returns ready-to-use database connection
func NewDBStorage(cfg config.DB) (*sql.DB, error) {
	if cfg.DSN == "" {
		return nil, errors.New("DSN is required")
	}

	connection, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	if err = connection.Ping(); err != nil {
		return nil, err
	}

	if err := runMigrations(connection, cfg); err != nil {
		connection.Close()
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	return connection, nil
}

// runMigrations executes database migrations for PostgreSQL.
//
// db: Established database connection
// cfg: Database configuration containing migration path
//
// Returns:
//   - error: Migration execution failure
//
// Migration behavior:
//   - If MigrationsPath is empty, skips migrations
//   - Uses file-based migrations from specified directory
//   - Only applies pending migrations (idempotent)
//   - Logs successful migration completion
func runMigrations(db *sql.DB, cfg config.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	migrationsPath := "file://migrations"
	if cfg.MigrationsPath != "" {
		migrationsPath = "file://" + cfg.MigrationsPath
	} else {
		return nil
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	slog.Info("Migrations finished", slog.String("filepath", migrationsPath))

	return nil
}
