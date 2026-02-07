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
	"os"
)

const aesKeyLen = 32

type (
	// Encryptor - interface for any encryptor
	Encryptor interface {
		Encrypt(plain []byte) ([]byte, error)
	}

	// X509Encryptor handles hybrid encryption of data.
	X509Encryptor struct {
		publicKey *rsa.PublicKey
	}
)

// NewX509Encryptor creates an X509Encryptor from X.509 certificate file.
//
// Certificate file must be in PEM format containing RSA public key.
func NewX509Encryptor(publicKeyPath string) (*X509Encryptor, error) {
	certificateBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	certificatePemBlock, _ := pem.Decode(certificateBytes)
	if certificatePemBlock == nil {
		return nil, errors.New("certificate not found")
	}
	certificate, err := x509.ParseCertificate(certificatePemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	if publicKey, ok := certificate.PublicKey.(*rsa.PublicKey); ok {
		return &X509Encryptor{publicKey: publicKey}, nil
	}

	return nil, errors.New("invalid public key type")
}

// Encrypt encrypts data using hybrid RSA/AES-GCM scheme.
//
// Output format: [2-byte key length][RSA-encrypted AES key][nonce][AES-GCM ciphertext]
func (e *X509Encryptor) Encrypt(plain []byte) ([]byte, error) {
	aesKey := make([]byte, aesKeyLen)
	if _, err := rand.Read(aesKey); err != nil {
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

	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	ciphertext := gcm.Seal(nil, nonce, plain, nil)

	encKey, err := rsa.EncryptPKCS1v15(rand.Reader, e.publicKey, aesKey)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(len(encKey)))
	buf.Write(encKey)
	buf.Write(nonce)
	buf.Write(ciphertext)
	return buf.Bytes(), nil
}
