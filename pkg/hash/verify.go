package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

type (
	Verifier interface {
		Verify(data []byte, sign string) bool
	}

	SHA256Verifier struct {
		key []byte
	}
)

func NewSHA256Verifier(key string) Verifier {
	return &SHA256Verifier{
		key: []byte(key),
	}
}

func (verifier *SHA256Verifier) Verify(data []byte, sign string) bool {
	h := hmac.New(sha256.New, verifier.key)
	h.Write(data)
	result := h.Sum(nil)

	signRaw, err := base64.StdEncoding.DecodeString(sign)

	if err != nil {
		return false
	}

	return hmac.Equal(signRaw, result)
}
