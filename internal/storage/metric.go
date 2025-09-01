package storage

import (
	"sync"

	"github.com/gabkaclassic/metrics/internal/model"
)

type MemStorage struct {
	Metrics map[string]models.Metrics
	Mutex   sync.RWMutex
}

func NewMemStorage() *MemStorage {

	return &MemStorage{
		Metrics: make(map[string]models.Metrics),
		Mutex:   sync.RWMutex{},
	}
}
