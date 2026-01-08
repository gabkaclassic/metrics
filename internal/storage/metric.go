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
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gabkaclassic/metrics/internal/config"
	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
//   - *pgxpool.Pool: Established database pool of connections
//   - error: Connection or migration failure details
//
// The function:
//  1. Validates configuration (DSN is required)
//  2. Opens and verifies database connection
//  3. Runs database migrations if configured
//  4. Returns ready-to-use database connection
func NewDBStorage(ctx context.Context, cfg config.DB) (DB, error) {
	if cfg.DSN == "" {
		return nil, errors.New("DSN is required")
	}

	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, err
	}
	pool.Config().MaxConns = int32(cfg.MaxConns)
	pool.Config().MaxConnLifetime = cfg.MaxConnTTL

	if err = pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	if err := runMigrations(cfg); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrations failed: %w", err)
	}
	return pool, nil
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
func runMigrations(cfg config.DB) error {
	if cfg.MigrationsPath == "" {
		return nil
	}

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to open db for migrations: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	migrationsPath := "file://migrations"
	if cfg.MigrationsPath != "" {
		migrationsPath = "file://" + cfg.MigrationsPath
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

type DB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Close()
}
