package validator

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance with custom password and PIN validators.
func New() *Validator {
	v := &Validator{
		validate: validator.New(),
	}

	// Register custom password strength validator
	v.validate.RegisterValidation("password_strength", passwordStrengthValidator)

	// Register custom no repeated chars validator
	v.validate.RegisterValidation("no_repeated", noRepeatedCharsValidator)

	// Register custom mixed characters validator
	v.validate.RegisterValidation("mixed_chars", mixedCharsValidator)

	// Register custom PIN validator
	v.validate.RegisterValidation("pin", pinValidator)

	return v
}

// passwordStrengthValidator checks for uppercase, lowercase, digit, and special char.
func passwordStrengthValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case isSpecialChar(ch):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// noRepeatedCharsValidator checks for 3+ repeated characters.
func noRepeatedCharsValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	count := 1

	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			count++
			if count >= 3 {
				return false
			}
		} else {
			count = 1
		}
	}
	return true
}

// mixedCharsValidator checks that password has both letters and non-letters.
func mixedCharsValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	hasLetter := false
	hasNonLetter := false

	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		} else {
			hasNonLetter = true
		}
	}

	return hasLetter && hasNonLetter
}

// isSpecialChar checks if a rune is a special character.
func isSpecialChar(r rune) bool {
	specialChars := regexp.MustCompile(`[!@#$%^&*()_+-=[]{}|;':",./<>?]`)
	return specialChars.MatchString(string(r))
}

// isCommonPINPattern checks for common PIN patterns (sequential, repeated, date).
func isCommonPINPattern(pin string) bool {
	// Sequential numbers
	sequential := []string{
		"123456", "234567", "345678", "456789",
		"654321", "765432", "876543", "987654",
	}

	// Repeated digits (all same)
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

	// Sequential patterns
	for _, seq := range sequential {
		if pin == seq {
			return true
		}
	}

	// Alternating patterns
	nums := make([]int, len(pin))
	for i, c := range pin {
		nums[i] = int(c - '0')
	}
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

	// Date patterns (MMDD)
	month := int(pin[0]-'0')*10 + int(pin[1]-'0')
	day := int(pin[2]-'0')*10 + int(pin[3]-'0')
	if month >= 1 && month <= 12 && day >= 1 && day <= 31 {
		return true
	}

	return false
}

// pinValidator validates PIN format (6 digits, no common patterns).
func pinValidator(fl validator.FieldLevel) bool {
	pin := fl.Field().String()

	// Must be exactly 6 digits
	if len(pin) != 6 {
		return false
	}

	// Must contain only digits
	for _, c := range pin {
		if c < '0' || c > '9' {
			return false
		}
	}

	// Check for common patterns
	return !isCommonPINPattern(pin)
}

// Struct validates a struct and returns a ValidationErrors map.
func (v *Validator) Struct(s any) error {
	return v.validate.Struct(s)
}

// ValidationErrors returns a readable error message for validation errors.
func ValidationErrors(err error) string {
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err.Error()
	}

	validationErrors := err.(validator.ValidationErrors)
	if len(validationErrors) == 0 {
		return ""
	}

	// Return first error for simplicity
	for _, e := range validationErrors {
		return formatFieldError(e)
	}
	return "validation failed"
}

// formatFieldError formats a single validation error.
func formatFieldError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()

	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return field + " must be at least " + e.Param() + " characters"
	case "max":
		return field + " must be at most " + e.Param() + " characters"
	case "len":
		return field + " must be exactly " + e.Param() + " characters"
	case "oneof":
		return field + " must be one of: " + e.Param()
	case "password_strength":
		return field + " must contain uppercase, lowercase, digit, and special character"
	case "no_repeated":
		return field + " must not contain 3 or more repeated characters"
	case "mixed_chars":
		return field + " must contain both letters and other characters"
	case "numeric":
		return field + " must contain only digits"
	case "uri":
		return field + " must be a valid URL"
	case "pin":
		return field + " must be a valid 6-digit PIN (no common patterns)"
	default:
		return field + " is invalid (" + tag + ")"
	}
}
