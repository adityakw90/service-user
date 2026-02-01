package security

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

// Argon2Hasher implements port.Hasher using the alexedwards/argon2id library.
type Argon2Hasher struct {
	params *argon2id.Params
}

// DefaultArgon2Params returns the recommended Argon2id parameters.
// These are suitable values for password hashing in production.
func DefaultArgon2Params() *argon2id.Params {
	return &argon2id.Params{
		Memory:      128 * 1024, // 128 MiB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// NewArgon2Hasher creates a new Argon2Hasher with custom parameters.
func NewArgon2Hasher(options *map[string]any) *Argon2Hasher {
	params := DefaultArgon2Params()
	if options != nil {
		for k, v := range *options {
			switch k {
			case "Memory":
				if n, ok := toUint32(v); ok {
					params.Memory = n
				}
			case "Iterations":
				if n, ok := toUint32(v); ok {
					params.Iterations = n
				}
			case "Parallelism":
				if n, ok := toUint8(v); ok {
					params.Parallelism = n
				}
			case "SaltLength":
				if n, ok := toUint32(v); ok {
					params.SaltLength = n
				}
			case "KeyLength":
				if n, ok := toUint32(v); ok {
					params.KeyLength = n
				}
			}
		}
	}
	return &Argon2Hasher{params: params}
}

// toUint32 converts various integer types to uint32.
func toUint32(v any) (uint32, bool) {
	switch val := v.(type) {
	case int:
		return uint32(val), true
	case int32:
		return uint32(val), true
	case int64:
		return uint32(val), true
	case uint:
		return uint32(val), true
	case uint32:
		return val, true
	case uint64:
		return uint32(val), true
	default:
		return 0, false
	}
}

// toUint8 converts various integer types to uint8.
func toUint8(v any) (uint8, bool) {
	switch val := v.(type) {
	case int:
		return uint8(val), true
	case int32:
		return uint8(val), true
	case int64:
		return uint8(val), true
	case uint:
		return uint8(val), true
	case uint8:
		return val, true
	case uint64:
		return uint8(val), true
	default:
		return 0, false
	}
}

// Hash creates a secure hash of the password using Argon2id.
func (h *Argon2Hasher) Hash(plain string) (string, error) {
	hash, err := argon2id.CreateHash(plain, h.params)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return hash, nil
}

// Compare compares a hashed password with a plain password.
func (h *Argon2Hasher) Compare(hashed, plain string) bool {
	match, err := argon2id.ComparePasswordAndHash(plain, hashed)
	if err != nil {
		return false
	}
	return match
}
