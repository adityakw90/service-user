package security

import (
	"fmt"

	"github.com/adityakw90/service-user/pkg/util"
	"golang.org/x/crypto/bcrypt"
)

// BCryptHasher implements PasswordHasher and PINHasher using bcrypt.
type BCryptHasher struct {
	cost int
}

// NewBCryptHasher creates a new BCryptHasher with the given cost.
// If cost is less than MinCost, DefaultCost is used.
func NewBCryptHasher(cost *int) *BCryptHasher {
	if cost == nil {
		cost = util.Ptr(bcrypt.DefaultCost)
	}
	if *cost < bcrypt.MinCost {
		*cost = bcrypt.DefaultCost
	}
	return &BCryptHasher{
		cost: *cost,
	}
}

// Hash generates a bcrypt hash of the password.
func (h *BCryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// Compare compares a bcrypt hashed password with a plain password.
func (h *BCryptHasher) Compare(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
