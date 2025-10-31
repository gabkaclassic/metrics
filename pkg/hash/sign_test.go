package hash

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSHA256Signer_Sign(t *testing.T) {
	tests := []struct {
		name string
		key  string
		data []byte
	}{
		{
			name: "regular data",
			key:  "test-key",
			data: []byte("test-data"),
		},
		{
			name: "empty data",
			key:  "test-key",
			data: []byte{},
		},
		{
			name: "special characters in data",
			key:  "test-key",
			data: []byte("test@data#123"),
		},
		{
			name: "different key",
			key:  "different-key",
			data: []byte("test-data"),
		},
		{
			name: "long data",
			key:  "test-key",
			data: []byte("very-long-test-data-string-with-many-characters"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer := NewSHA256Signer(tt.key)
			result := signer.Sign(tt.data)

			assert.NotEmpty(t, result)
			assert.True(t, isValidBase64(result))
		})
	}
}

func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
