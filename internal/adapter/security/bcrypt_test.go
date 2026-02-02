package security

import (
	"strings"
	"testing"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
	"github.com/adityakw90/service-user/pkg/util"
	"golang.org/x/crypto/bcrypt"
)

func TestBCryptHasher_Hash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		cost     *int
		wantErr  bool
	}{
		{
			name:     "hashes simple password",
			password: "secure_password_123",
			cost:     nil,
			wantErr:  false,
		},
		{
			name:     "hashes short password",
			password: "123",
			cost:     nil,
			wantErr:  false,
		},
		{
			name:     "hashes long password (near limit)",
			password: strings.Repeat("a", 70),
			cost:     nil,
			wantErr:  false,
		},
		{
			name:     "hashes with custom cost",
			password: "test_password",
			cost:     util.Ptr(bcrypt.DefaultCost),
			wantErr:  false,
		},
		{
			name:     "hashes with high cost",
			password: "test_password",
			cost:     util.Ptr(12),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBCryptHasher(tt.cost)
			hash, err := hasher.Hash(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
					t.Errorf("hash should start with $2a$ or $2b$, got: %s", hash)
				}
				if hash == tt.password {
					t.Error("hash should not equal password")
				}
			}
		})
	}
}

func TestBCryptHasher_Compare(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wrongPass   string
		wantCorrect bool
		wantWrong   bool
		cost        *int
	}{
		{
			name:        "correct password matches, wrong does not",
			password:    "correct_password",
			wrongPass:   "wrong_password",
			wantCorrect: true,
			wantWrong:   false,
			cost:        nil,
		},
		{
			name:        "empty password comparison",
			password:    "test_password",
			wrongPass:   "",
			wantCorrect: true,
			wantWrong:   false,
			cost:        nil,
		},
		{
			name:        "with custom cost",
			password:    "test_password",
			wrongPass:   "wrong",
			wantCorrect: true,
			wantWrong:   false,
			cost:        util.Ptr(bcrypt.DefaultCost),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBCryptHasher(tt.cost)
			hash, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			// Compare correct password
			gotMatch := hasher.Compare(hash, tt.password)
			if gotMatch != tt.wantCorrect {
				t.Errorf("Compare() with correct password = %v, want %v", gotMatch, tt.wantCorrect)
			}

			// Compare wrong password
			gotWrongMatch := hasher.Compare(hash, tt.wrongPass)
			if gotWrongMatch != tt.wantWrong {
				t.Errorf("Compare() with wrong password = %v, want %v", gotWrongMatch, tt.wantWrong)
			}
		})
	}
}

func TestBCryptHasher_DifferentPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"different password cross-check", "password_one"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBCryptHasher(nil)
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

func TestBCryptHasher_SamePasswordDifferentHashes(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"same password produces different hashes", "same_password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBCryptHasher(nil)

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

func TestBCryptHasher_EmptyPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"empty password should not match hash", "test_password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBCryptHasher(nil)
			hash, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if hasher.Compare(hash, "") {
				t.Error("empty password should not match hash")
			}
		})
	}
}

func TestBCryptHasher_InvalidHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		password string
		want     bool
	}{
		{"invalid hash format", "invalid_hash", "password", false},
		{"empty hash", "", "password", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBCryptHasher(nil)
			got := hasher.Compare(tt.hash, tt.password)
			if got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBCryptHasher_MinCost(t *testing.T) {
	tests := []struct {
		name    string
		cost    int
		wantErr bool
	}{
		{"below min cost uses default", bcrypt.MinCost - 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBCryptHasher(&tt.cost)
			hash, err := hasher.Hash("test_password")
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !hasher.Compare(hash, "test_password") {
				t.Error("password should match its hash")
			}
		})
	}
}

func TestBCryptHasher_UnicodePassword(t *testing.T) {
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
			hasher := NewBCryptHasher(nil)
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

func TestBCryptHasher_ImplementsPortHasher(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"BCryptHasher implements port.Hasher"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var _ portsec.Hasher = (*BCryptHasher)(nil)
		})
	}
}
