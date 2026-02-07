package crypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewX509Encryptor(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError bool
	}{
		{
			name: "valid x509 certificate with rsa public key",
			setup: func(t *testing.T) string {
				priv, err := rsa.GenerateKey(rand.Reader, 2048)
				assert.NoError(t, err)

				template := &x509.Certificate{
					SerialNumber: big.NewInt(1),
					NotBefore:    time.Now().Add(-time.Hour),
					NotAfter:     time.Now().Add(time.Hour),
					KeyUsage:     x509.KeyUsageKeyEncipherment,
				}

				certDER, err := x509.CreateCertificate(
					rand.Reader,
					template,
					template,
					&priv.PublicKey,
					priv,
				)
				assert.NoError(t, err)

				certPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "CERTIFICATE",
					Bytes: certDER,
				})

				path := filepath.Join(t.TempDir(), "cert.pem")
				assert.NoError(t, os.WriteFile(path, certPEM, 0600))
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
			name: "pem but not certificate",
			setup: func(t *testing.T) string {
				priv, err := rsa.GenerateKey(rand.Reader, 2048)
				assert.NoError(t, err)

				keyPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "RSA PRIVATE KEY",
					Bytes: x509.MarshalPKCS1PrivateKey(priv),
				})

				path := filepath.Join(t.TempDir(), "key.pem")
				assert.NoError(t, os.WriteFile(path, keyPEM, 0600))
				return path
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			enc, err := NewX509Encryptor(path)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, enc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, enc)
				assert.NotNil(t, enc.publicKey)
			}
		})
	}
}

func TestEncrypt(t *testing.T) {
	tests := []struct {
		name        string
		plain       []byte
		expectError bool
	}{
		{
			name:        "encrypt non-empty payload",
			plain:       []byte("secret data"),
			expectError: false,
		},
		{
			name:        "encrypt empty payload",
			plain:       []byte{},
			expectError: false,
		},
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	encryptor := &X509Encryptor{
		publicKey: &priv.PublicKey,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := encryptor.Encrypt(tt.plain)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, encrypted)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, encrypted)
			assert.Greater(t, len(encrypted), 0)

			buf := bytes.NewReader(encrypted)

			var keyLen uint16
			err = binary.Read(buf, binary.BigEndian, &keyLen)
			assert.NoError(t, err)
			assert.Greater(t, int(keyLen), 0)

			encKey := make([]byte, keyLen)
			_, err = io.ReadFull(buf, encKey)
			assert.NoError(t, err)

			assert.Equal(t, 256, len(encKey))

			rest, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Greater(t, len(rest), 0)
		})
	}
}
