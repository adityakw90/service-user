package model

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
)

const (
	// PINLength is the fixed PIN length.
	PINLength = 6
	// PINMaxAttempts is the maximum failed PIN verification attempts.
	PINMaxAttempts = 5
)

// PIN represents a validated PIN value object.
type PIN struct {
	value string
}

// NewPIN creates a new PIN after validation.
func NewPIN(value string) (*PIN, error) {
	if err := validatePIN(value); err != nil {
		return nil, err
	}
	return &PIN{value: value}, nil
}

// String returns the PIN value.
func (p *PIN) String() string {
	return p.value
}

// Value returns the PIN value.
func (p *PIN) Value() string {
	return p.value
}

// Compare compares this PIN with another PIN value.
func (p *PIN) Compare(provided string) bool {
	providedPIN, err := NewPIN(provided)
	if err != nil {
		return false
	}
	return p.value == providedPIN.value
}

// validatePIN validates a 6-digit PIN.
func validatePIN(value string) error {
	if len(value) != PINLength {
		return errors.ErrInvalidPINLength
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(value) {
		return errors.ErrInvalidPINFormat
	}

	// Check for common weak patterns
	if isCommonPINPattern(value) {
		return errors.ErrPINContainsCommonPattern
	}

	return nil
}

// isCommonPINPattern checks for common PIN patterns.
func isCommonPINPattern(pin string) bool {
	// Sequential numbers
	sequential := []string{
		"123456", "234567", "345678", "456789",
		"654321", "765432", "876543", "987654",
	}

	// Repeated digits
	allSame := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return true
	}

	// Check for sequential patterns
	for _, seq := range sequential {
		if pin == seq {
			return true
		}
	}

	// Check for ascending/descending patterns
	nums := make([]int, 6)
	for i, c := range pin {
		nums[i], _ = strconv.Atoi(string(c))
	}

	// Check for alternating patterns
	alternating := true
	for i := 2; i < len(nums); i++ {
		if nums[i]-nums[i-1] != nums[i-1]-nums[i-2] {
			alternating = false
			break
		}
	}
	if alternating {
		return true
	}

	// Check for date patterns (MMDD)
	month, _ := strconv.Atoi(pin[:2])
	day, _ := strconv.Atoi(pin[2:])
	if (month >= 1 && month <= 12) && (day >= 1 && day <= 31) {
		return true
	}

	return false
}

// ValidateForAttempt validates PIN and tracks attempts.
func (p *PIN) ValidateForAttempt(provided string, attempts *int) (bool, error) {
	if *attempts >= PINMaxAttempts {
		return false, errors.ErrPINTooManyAttempts
	}

	*attempts++
	return p.Compare(provided), nil
}

// ContainsDuplicates checks if PIN has duplicate adjacent digits.
func (p *PIN) ContainsDuplicates() bool {
	for i := 1; i < len(p.value); i++ {
		if p.value[i] == p.value[i-1] {
			return true
		}
	}
	return false
}

// HasOnlyEvenOrOddDigits checks if all digits are even or all are odd.
func (p *PIN) HasOnlyEvenOrOddDigits() bool {
	firstIsEven := (p.value[0]-'0')%2 == 0
	for i := 1; i < len(p.value); i++ {
		digit := p.value[i] - '0'
		isEven := digit%2 == 0
		if isEven != firstIsEven {
			return false
		}
	}
	return true
}

// ToInt converts PIN to integer (for storage if needed).
func (p *PIN) ToInt() (int, error) {
	return strconv.Atoi(p.value)
}

// FromInt creates PIN from integer.
func FromInt(n int) (*PIN, error) {
	s := strconv.Itoa(n)
	// Pad with leading zeros if needed
	for len(s) < PINLength {
		s = "0" + s
	}
	// Truncate if too long
	if len(s) > PINLength {
		s = s[len(s)-PINLength:]
	}
	return NewPIN(s)
}

// Split returns PIN as two 3-digit parts for additional validation.
func (p *PIN) Split() (first, second string) {
	return p.value[:3], p.value[3:]
}

// Sum returns the sum of all digits.
func (p *PIN) Sum() int {
	sum := 0
	for _, c := range p.value {
		sum += int(c - '0')
	}
	return sum
}

// Contains checks if PIN contains the given substring.
func (p *PIN) Contains(substr string) bool {
	return strings.Contains(p.value, substr)
}
