package security

import (
	"github.com/google/uuid"
)

// UIDGenerator generates UUID v7 identifiers.
type UIDGenerator struct{}

// NewUIDGenerator creates a new UIDGenerator.
func NewUIDGenerator() *UIDGenerator {
	return &UIDGenerator{}
}

// NewV7 generates a new UUID v7 using the google/uuid library.
func (g *UIDGenerator) New() string {
	u, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return u.String()
}
