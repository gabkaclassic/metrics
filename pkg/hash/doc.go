// Package hash provides cryptographic signing and verification for data integrity.
//
// The package implements HMAC-SHA256 based signing and verification for ensuring
// data authenticity and integrity in distributed systems. It provides:
//   - Signer: Creates cryptographic signatures for data
//   - Verifier: Validates signatures against data
//
// Both use base64 encoding for signature representation, making them suitable
// for HTTP headers and other text-based protocols.
package hash
