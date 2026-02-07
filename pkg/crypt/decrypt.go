package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"io"
	"os"
)

const gcmSize = 12

type (

	// Decryptor - interface for any decryptor
	Decryptor interface {
		Decrypt(payload []byte) ([]byte, error)
	}

	// X509Decryptor handles decryption of hybrid-encrypted data.
	X509Decryptor struct {
		privateKey *rsa.PrivateKey
	}
)

// NewX509Decryptor creates a X509Decryptor from RSA private key file.
//
// Key file must be in PEM format with PKCS#1 encoding.
func NewX509Decryptor(privateKeyPath string) (*X509Decryptor, error) {

	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	privateKeyPemBlock, _ := pem.Decode(keyBytes)
	if privateKeyPemBlock == nil {
		return nil, errors.New("private pem key not found")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return &X509Decryptor{
		privateKey: privateKey,
	}, nil
}

// Decrypt decrypts hybrid-encrypted payload.
//
// Expected format: [2-byte key length][RSA-encrypted AES key][12-byte nonce][AES-GCM ciphertext]
func (d *X509Decryptor) Decrypt(payload []byte) ([]byte, error) {
	buf := bytes.NewReader(payload)

	var keyLen uint16
	if err := binary.Read(buf, binary.BigEndian, &keyLen); err != nil {
		return nil, err
	}

	encKey := make([]byte, keyLen)
	if _, err := io.ReadFull(buf, encKey); err != nil {
		return nil, err
	}

	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, d.privateKey, encKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcmSize)
	if _, err := io.ReadFull(buf, nonce); err != nil {
		return nil, err
	}

	ciphertext, err := io.ReadAll(buf)

	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plain, nil
}
