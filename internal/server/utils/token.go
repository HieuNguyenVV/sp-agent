package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken returns a deterministic SHA-256 hex digest of token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
