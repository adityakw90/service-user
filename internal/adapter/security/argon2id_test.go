package security

import (
	"strings"
	"testing"

	"github.com/adityakw90/service-user/internal/core/port"
)

func TestArgon2Hasher_Hash(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantFn    func(string) bool
	}{
		{
			name:     "hashes simple password",
			password: "test_password_123",
			wantFn: func(hash string) bool {
				return strings.HasPrefix(hash, "$argon2id$") && hash != "test_password_123"
			},
		},
		{
			name:     "hashes empty password",
			password: "",
			wantFn: func(hash string) bool {
				return strings.HasPrefix(hash, "$argon2id$") && hash != ""
			},
		},
		{
			name:     "hashes short password",
			password: "123",
			wantFn: func(hash string) bool {
				return strings.HasPrefix(hash, "$argon2id$") && hash != "123"
			},
		},
		{
			name:     "same password produces different hashes",
			password: "same_password",
			wantFn: func(hash string) bool {
				return strings.HasPrefix(hash, "$argon2id$")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewArgon2Hasher(nil)
			hash, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			if !tt.wantFn(hash) {
				t.Errorf("Hash() = %v, want valid argon2id hash", hash)
			}
		})
	}
}

func TestArgon2Hasher_Compare(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wrongPass   string
		wantCorrect bool
		wantWrong   bool
	}{
		{
			name:        "correct password matches, wrong does not",
			password:    "my_secure_password",
			wrongPass:   "wrong_password",
			wantCorrect: true,
			wantWrong:   false,
		},
		{
			name:        "empty password comparison",
			password:    "test_password",
			wrongPass:   "",
			wantCorrect: true,
			wantWrong:   false,
		},
		{
			name:        "special characters in password",
			password:    "p@$$w0rd!#",
			wrongPass:   "wrong",
			wantCorrect: true,
			wantWrong:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewArgon2Hasher(nil)
			hash, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			gotMatch := hasher.Compare(hash, tt.password)
			if gotMatch != tt.wantCorrect {
				t.Errorf("Compare() with correct password = %v, want %v", gotMatch, tt.wantCorrect)
			}

			gotWrongMatch := hasher.Compare(hash, tt.wrongPass)
			if gotWrongMatch != tt.wantWrong {
				t.Errorf("Compare() with wrong password = %v, want %v", gotWrongMatch, tt.wantWrong)
			}
		})
	}
}

func TestArgon2Hasher_DifferentPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"different passwords", "password_one"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewArgon2Hasher(nil)
			password1 := "password_one"
			password2 := "password_two"

			hash1, err := hasher.Hash(password1)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			hash2, err := hasher.Hash(password2)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if hasher.Compare(hash1, password2) {
				t.Error("password2 should not match hash1")
			}
			if hasher.Compare(hash2, password1) {
				t.Error("password1 should not match hash2")
			}
		})
	}
}

func TestArgon2Hasher_DefaultParams(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"default params work", "test_password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewArgon2Hasher(nil)
			hash, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			if !hasher.Compare(hash, tt.password) {
				t.Error("password should match its hash")
			}
		})
	}
}

func TestArgon2Hasher_LongPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"long password (100 chars)", strings.Repeat("a", 100)},
		{"very long password (1000 chars)", strings.Repeat("b", 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewArgon2Hasher(nil)
			hash, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			if !hasher.Compare(hash, tt.password) {
				t.Error("long password should match its hash")
			}
		})
	}
}

func TestArgon2Hasher_UnicodePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"cyrillic password", "пароль"},
		{"chinese password", "密码"},
		{"emoji password", "🔐🔑"},
		{"mixed unicode password", "пароль_密码_🔐"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewArgon2Hasher(nil)
			hash, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			if !hasher.Compare(hash, tt.password) {
				t.Error("unicode password should match its hash")
			}
		})
	}
}

func TestArgon2Hasher_InvalidHash(t *testing.T) {
	tests := []struct {
		name      string
		hash      string
		password  string
		wantMatch bool
	}{
		{"invalid hash format", "invalid_hash", "password", false},
		{"malformed hash", "$argon2id$malformed", "password", false},
		{"empty hash", "", "password", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewArgon2Hasher(nil)
			got := hasher.Compare(tt.hash, tt.password)
			if got != tt.wantMatch {
				t.Errorf("Compare() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestArgon2Hasher_SamePasswordDifferentHashes(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"same password produces different hashes", "same_password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewArgon2Hasher(nil)
			hash1, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			hash2, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if hash1 == hash2 {
				t.Error("same password should produce different hashes (different salt)")
			}
			if !hasher.Compare(hash1, tt.password) {
				t.Error("hash1 should match password")
			}
			if !hasher.Compare(hash2, tt.password) {
				t.Error("hash2 should match password")
			}
		})
	}
}

func TestArgon2Hasher_ImplementsPortHasher(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Argon2Hasher implements port.Hasher"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var _ port.Hasher = (*Argon2Hasher)(nil)
		})
	}
}
