package security

import (
	"fmt"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
)

// NewHasher creates a new hasher based on the type.
// Supported types: "argon2" (default), "bcrypt", "sha256"
func NewHasher(hasherType string, params map[string]any) (portsec.Hasher, error) {
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
