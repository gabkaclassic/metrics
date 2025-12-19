package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.Metrics)
	assert.Empty(t, storage.Metrics)
}
