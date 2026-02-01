package security

import (
	"fmt"

	"github.com/adityakw90/service-user/internal/core/port"
	"github.com/adityakw90/service-user/pkg/util"
)

// NewHasher creates a new hasher based on the type.
// Supported types: "argon2" (default), "bcrypt", "sha256"
func NewHasher(hasherType string, params map[string]any) (port.Hasher, error) {
	switch hasherType {
	case "argon2", "":
		return NewArgon2Hasher(&params), nil
	case "bcrypt":
		cost, _ := params["cost"].(int)
		return NewBCryptHasher(&cost), nil
	case "sha256":
		salt, _ := params["salt"].(string)
		return NewSHA256Hasher(&salt), nil
	default:
		return nil, fmt.Errorf("unknown hasher type: %s", hasherType)
	}
}

// NewPINHasher creates a hasher optimized for PIN hashing.
// PINs are shorter and typically numeric, so we use stronger parameters.
func NewPINHasher(hasherType string) (port.Hasher, error) {
	switch hasherType {
	case "argon2", "":
		return NewArgon2Hasher(&map[string]any{
			"Iterations": 4, // Higher iterations for shorter PINs
		}), nil
	case "bcrypt":
		return NewBCryptHasher(util.Ptr(14)), nil // Higher cost for PINs
	case "sha256":
		return NewSHA256Hasher(nil), nil
	default:
		return nil, fmt.Errorf("unknown hasher type: %s", hasherType)
	}
}
