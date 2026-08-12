package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandomPassword(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "length too short",
			length:  3,
			wantErr: true,
		},
		{
			name:    "minimum length",
			length:  4,
			wantErr: false,
		},
		{
			name:    "standard length",
			length:  16,
			wantErr: false,
		},
		{
			name:    "long length",
			length:  64,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GenerateRandomPassword(tt.length)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, password)
				return
			}

			require.NoError(t, err)
			assert.Len(t, password, tt.length)

			hasUpper := false
			hasLower := false
			hasDigit := false
			hasSpecial := false

			upperCase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
			lowerCase := "abcdefghijklmnopqrstuvwxyz"
			digits := "0123456789"
			specialChars := "!@#$%^&*()-_=+[]{}|;:,.<>?/"

			for i := 0; i < len(password); i++ {
				char := string(password[i])
				if strings.Contains(upperCase, char) {
					hasUpper = true
				} else if strings.Contains(lowerCase, char) {
					hasLower = true
				} else if strings.Contains(digits, char) {
					hasDigit = true
				} else if strings.Contains(specialChars, char) {
					hasSpecial = true
				}
			}

			assert.True(t, hasUpper, "password should contain at least one uppercase letter")
			assert.True(t, hasLower, "password should contain at least one lowercase letter")
			assert.True(t, hasDigit, "password should contain at least one digit")
			assert.True(t, hasSpecial, "password should contain at least one special character")
		})
	}

	// Verify randomness (uniqueness)
	t.Run("randomness verification", func(t *testing.T) {
		passwords := make(map[string]bool)
		for range 50 {
			pw, err := GenerateRandomPassword(16)
			require.NoError(t, err)
			assert.False(t, passwords[pw], "Generated password should be unique")
			passwords[pw] = true
		}
	})
}
