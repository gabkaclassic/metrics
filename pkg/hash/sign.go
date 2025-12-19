package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

type (
	// Signer defines the interface for creating cryptographic signatures.
	// Implementations produce signatures that can be later verified to ensure
	// data authenticity and integrity.
	Signer interface {
		// Sign creates a cryptographic signature for the given data.
		// data: Raw bytes to sign.
		// Returns: encoded signature string.
		Sign(data []byte) string
	}

	// SHA256Signer implements Signer using HMAC-SHA256 algorithm.
	// Provides strong cryptographic signing suitable for security-sensitive applications.
	SHA256Signer struct {
		key []byte
	}
)

// NewSHA256Signer creates a new HMAC-SHA256 signer with the provided key.
//
// key: Secret key used for signing. For security, should be:
//   - At least 32 bytes (256 bits) for SHA256
//   - Generated cryptographically randomly
//   - Stored securely (not in source code)
//
// Returns: Initialized Signer ready to create signatures.
// The same key must be used by the verifier to validate signatures.
func NewSHA256Signer(key string) Signer {
	return &SHA256Signer{
		key: []byte(key),
	}
}

// Sign creates an HMAC-SHA256 signature for the given data.
//
// data: Bytes to sign. Can be any length; HMAC handles hashing internally.
//
// Returns: Base64-encoded signature string
//
// The signature is deterministic: same data + same key = same signature.
// Changing any byte in data or using different key produces different signature.
func (encoder *SHA256Signer) Sign(data []byte) string {
	h := hmac.New(sha256.New, encoder.key)
	h.Write(data)
	result := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(result)
}
