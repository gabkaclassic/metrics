package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

type (
	Signer interface {
		Sign(data []byte) string
	}

	SHA256Signer struct {
		key []byte
	}
)

func NewSHA256Signer(key string) Signer {
	return &SHA256Signer{
		key: []byte(key),
	}
}

func (encoder *SHA256Signer) Sign(data []byte) string {
	h := hmac.New(sha256.New, encoder.key)
	h.Write(data)
	result := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(result)
}
