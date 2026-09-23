package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// newEnrollSecret returns a one-time plaintext token and its SHA-256 hash.
func newEnrollSecret() (raw, hash string) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("api: rand: %v", err))
	}
	raw = "mhk_" + hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	return raw, hex.EncodeToString(sum[:])
}

// hashSecret returns the SHA-256 hex digest of a secret.
func hashSecret(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
