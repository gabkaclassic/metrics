package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

type (
	// Verifier defines the interface for validating cryptographic signatures.
	// Implementations check that data matches the provided signature,
	// ensuring it hasn't been tampered with and originated from trusted source.
	Verifier interface {
		// Verify checks if the provided signature matches the given data.
		// data: Raw bytes that were supposedly signed.
		// sign: encoded signature to verify.
		// Returns: true if signature is valid, false otherwise.
		Verify(data []byte, sign string) bool
	}

	// SHA256Verifier implements Verifier using HMAC-SHA256 algorithm.
	// Validates signatures created by SHA256Signer with the same secret key.
	SHA256Verifier struct {
		key []byte
	}
)

// NewSHA256Verifier creates a new HMAC-SHA256 verifier with the provided key.
//
// key: Secret key used for verification. Must be identical to the key
//
//	used by the Signer that created the signatures.
//
// Returns: Initialized Verifier ready to validate signatures.
func NewSHA256Verifier(key string) Verifier {
	return &SHA256Verifier{
		key: []byte(key),
	}
}

// Verify checks if the provided signature is valid for the given data.
//
// data: Original data bytes that were signed.
// sign: Base64-encoded signature to verify.
//
// Returns:
//   - true: Signature is valid (data authentic and untampered)
//   - false: Signature is invalid (tampering, wrong key, or malformed signature)
//
// The verification uses constant-time comparison to prevent timing attacks.
// Invalid base64 encoding in signature also returns false.
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
