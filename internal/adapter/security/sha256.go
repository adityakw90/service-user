package security

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/adityakw90/service-user/pkg/util"
)

// SHA256Hasher implements port.Hasher using SHA256.
// WARNING: SHA256 is not suitable for password hashing in production.
// It is extremely fast and vulnerable to brute-force attacks.
// Only use this for non-security-critical hashing or when required by legacy systems.
type SHA256Hasher struct {
	salt string // Optional salt for additional protection
}

// NewSHA256Hasher creates a new SHA256Hasher.
func NewSHA256Hasher(salt *string) *SHA256Hasher {
	if salt == nil {
		salt = util.Ptr("")
	}
	return &SHA256Hasher{salt: *salt}
}

// Hash creates a SHA256 hash of the input with optional salt.
func (h *SHA256Hasher) Hash(plain string) (string, error) {
	data := plain
	if h.salt != "" {
		data = h.salt + plain
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

// Compare compares a SHA256 hashed value with a plain value.
func (h *SHA256Hasher) Compare(hashed, plain string) bool {
	newHash, err := h.Hash(plain)
	if err != nil {
		return false
	}
	return hashed == newHash
}
