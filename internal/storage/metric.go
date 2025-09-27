package storage

import (
	"github.com/gabkaclassic/metrics/internal/model"
)

type MemStorage struct {
	Metrics map[string]models.Metrics
}

func NewMemStorage() *MemStorage {

	return &MemStorage{
		Metrics: make(map[string]models.Metrics),
	}
}
