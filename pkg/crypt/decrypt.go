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

type Decryptor struct {
	privateKey *rsa.PrivateKey
}

func NewDecryptor(privateKeyPath string) (*Decryptor, error) {

	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	privateKeyPemBlock, _ := pem.Decode(keyBytes)
	if privateKeyPemBlock == nil {
		return nil, errors.New("certificate not found")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return &Decryptor{
		privateKey: privateKey,
	}, nil
}

func (d *Decryptor) Decrypt(payload []byte) ([]byte, error) {
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

	ciphertext, _ := io.ReadAll(buf)
	block, _ := aes.NewCipher(aesKey)
	gcm, _ := cipher.NewGCM(block)
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plain, nil
}
