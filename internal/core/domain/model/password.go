package model

import (
	"strings"
	"time"
	"unicode"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
)

// Password represents a validated password value object.
type Password struct {
	value string
}

// PasswordRules defines password validation rules.
type PasswordRules struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// NewPasswordRules creates default password validation rules.
func NewPasswordRules() *PasswordRules {
	return &PasswordRules{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
	}
}

// NewPassword creates a new Password after validation.
func NewPassword(value string) (*Password, error) {
	rules := NewPasswordRules()
	if err := rules.Validate(value); err != nil {
		return nil, err
	}
	return &Password{value: value}, nil
}

// NewPasswordWithRules creates a password with custom validation rules.
func NewPasswordWithRules(value string, rules *PasswordRules) (*Password, error) {
	if err := rules.Validate(value); err != nil {
		return nil, err
	}
	return &Password{value: value}, nil
}

// String returns the password value.
func (p *Password) String() string {
	return p.value
}

// Value returns the password value.
func (p *Password) Value() string {
	return p.value
}

// Validate performs password validation.
func (r *PasswordRules) Validate(value string) error {
	if len(value) < r.MinLength {
		return errors.ErrPasswordTooShort
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, ch := range value {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", ch):
			hasSpecial = true
		}
	}

	if r.RequireUpper && !hasUpper {
		return errors.ErrPasswordNoUppercase
	}
	if r.RequireLower && !hasLower {
		return errors.ErrPasswordNoLowercase
	}
	if r.RequireDigit && !hasDigit {
		return errors.ErrPasswordNoDigit
	}
	if r.RequireSpecial && !hasSpecial {
		return errors.ErrPasswordNoSpecial
	}

	// Check for common weak patterns
	if hasRepeatedChars(value, 3) {
		return errors.ErrPasswordWeakPattern
	}

	if isOnlyNumbersOrLetters(value) {
		return errors.ErrPasswordWeakPattern
	}

	return nil
}

// hasRepeatedChars checks for repeated character sequences.
func hasRepeatedChars(s string, maxRepeat int) bool {
	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= maxRepeat {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

// isOnlyNumbersOrLetters checks if password contains only numbers or letters.
func isOnlyNumbersOrLetters(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsExpired checks if password was changed before the threshold.
func (p *Password) IsExpired(maxAge time.Duration, changedAt time.Time) bool {
	return time.Since(changedAt) > maxAge
}
