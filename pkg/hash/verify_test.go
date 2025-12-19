package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSHA256Verifier_Verify(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		data     []byte
		sign     string
		expected bool
	}{
		{
			name:     "valid signature",
			key:      "test-key",
			data:     []byte("test-data"),
			sign:     "IaKG/W/Z9SZ2AHxm0PiD20bQYVjCZtM/tTfCO8YY5Wc=",
			expected: true,
		},
		{
			name:     "invalid signature",
			key:      "test-key",
			data:     []byte("test-data"),
			sign:     "invalid-signature",
			expected: false,
		},
		{
			name:     "empty data",
			key:      "test-key",
			data:     []byte{},
			sign:     "JxHMI+mrG4qbwP6ZEjjakmcWJKnr2vHBq+wG5+mhT5s=",
			expected: true,
		},
		{
			name:     "wrong key",
			key:      "wrong-key",
			data:     []byte("test-data"),
			sign:     "pXNY6Vs2c0dM7sBsXW6bQ3X6WJPSqcbql1k7p3G0n/g=",
			expected: false,
		},
		{
			name:     "malformed base64",
			key:      "test-key",
			data:     []byte("test-data"),
			sign:     "!!!malformed!!!",
			expected: false,
		},
		{
			name:     "different data",
			key:      "test-key",
			data:     []byte("different-data"),
			sign:     "pXNY6Vs2c0dM7sBsXW6bQ3X6WJPSqcbql1k7p3G0n/g=",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := NewSHA256Verifier(tt.key)
			result := verifier.Verify(tt.data, tt.sign)
			assert.Equal(t, tt.expected, result)
		})
	}
}
