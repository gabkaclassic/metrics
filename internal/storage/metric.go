package storage

import (
	"database/sql"
	"fmt"
	"github.com/gabkaclassic/metrics/internal/config"
	"github.com/gabkaclassic/metrics/internal/model"
	_ "github.com/lib/pq"
)

type MemStorage struct {
	Metrics map[string]models.Metrics
}

func NewMemStorage() *MemStorage {

	return &MemStorage{
		Metrics: make(map[string]models.Metrics),
	}
}

func NewDBStorage(cfg config.DB) (*sql.DB, error) {
	var connectionString string

	if len(cfg.DSN) > 0 {
		connectionString = cfg.DSN
	} else {
		connectionString = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSL,
		)
	}
	connection, err := sql.Open(
		cfg.Driver,
		connectionString,
	)

	if err != nil {
		return nil, err
	}

	err = connection.Ping()

	if err != nil {
		return nil, err
	}

	return connection, nil
}
