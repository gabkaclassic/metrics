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

type Encryptor struct {
	publicKey *rsa.PublicKey
}

func NewEncryptor(publicKeyPath string) (*Encryptor, error) {
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
	publicKey := certificate.PublicKey.(*rsa.PublicKey)
	return &Encryptor{publicKey: publicKey}, nil
}

func (e *Encryptor) Encrypt(plain []byte) ([]byte, error) {
	aesKey := make([]byte, aesKeyLen)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, err
	}

	block, _ := aes.NewCipher(aesKey)
	gcm, _ := cipher.NewGCM(block)
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
