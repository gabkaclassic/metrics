package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.Metrics)
	assert.Empty(t, storage.Metrics)
}
