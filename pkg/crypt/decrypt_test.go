package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewX509Decryptor(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError bool
	}{
		{
			name: "valid pkcs1 private key",
			setup: func(t *testing.T) string {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				assert.NoError(t, err)

				pemBytes := pem.EncodeToMemory(&pem.Block{
					Type:  "RSA PRIVATE KEY",
					Bytes: x509.MarshalPKCS1PrivateKey(key),
				})

				path := filepath.Join(t.TempDir(), "key.pem")
				assert.NoError(t, os.WriteFile(path, pemBytes, 0600))
				return path
			},
			expectError: false,
		},
		{
			name: "file not found",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing.pem")
			},
			expectError: true,
		},
		{
			name: "not a pem file",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "bad.pem")
				assert.NoError(t, os.WriteFile(path, []byte("garbage"), 0600))
				return path
			},
			expectError: true,
		},
		{
			name: "pem but not pkcs1 key",
			setup: func(t *testing.T) string {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				assert.NoError(t, err)

				pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
				assert.NoError(t, err)

				pemBytes := pem.EncodeToMemory(&pem.Block{
					Type:  "PRIVATE KEY",
					Bytes: pkcs8,
				})

				path := filepath.Join(t.TempDir(), "pkcs8.pem")
				assert.NoError(t, os.WriteFile(path, pemBytes, 0600))
				return path
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			dec, err := NewX509Decryptor(path)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, dec)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dec)
				assert.NotNil(t, dec.privateKey)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError bool
	}{
		{
			name: "valid pkcs1 private key",
			setup: func(t *testing.T) string {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				assert.NoError(t, err)

				pemBytes := pem.EncodeToMemory(&pem.Block{
					Type:  "RSA PRIVATE KEY",
					Bytes: x509.MarshalPKCS1PrivateKey(key),
				})

				path := filepath.Join(t.TempDir(), "key.pem")
				assert.NoError(t, os.WriteFile(path, pemBytes, 0600))
				return path
			},
			expectError: false,
		},
		{
			name: "file not found",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing.pem")
			},
			expectError: true,
		},
		{
			name: "not a pem file",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "bad.pem")
				assert.NoError(t, os.WriteFile(path, []byte("garbage"), 0600))
				return path
			},
			expectError: true,
		},
		{
			name: "pem but not pkcs1 key",
			setup: func(t *testing.T) string {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				assert.NoError(t, err)

				pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
				assert.NoError(t, err)

				pemBytes := pem.EncodeToMemory(&pem.Block{
					Type:  "PRIVATE KEY",
					Bytes: pkcs8,
				})

				path := filepath.Join(t.TempDir(), "pkcs8.pem")
				assert.NoError(t, os.WriteFile(path, pemBytes, 0600))
				return path
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			dec, err := NewX509Decryptor(path)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, dec)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dec)
				assert.NotNil(t, dec.privateKey)
			}
		})
	}
}
