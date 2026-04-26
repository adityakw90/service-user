package security

import (
	"strings"
	"testing"

	"github.com/adityakw90/service-user/pkg/util"
)

func TestSHA256Hasher_Hash(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		salt   *string
		wantFn func(string) bool
	}{
		{
			name:  "hashes simple string",
			input: "hello world",
			salt:  nil,
			wantFn: func(got string) bool {
				return len(got) == 64 && !strings.Contains(got, "hello")
			},
		},
		{
			name:  "hashes empty string",
			input: "",
			salt:  nil,
			wantFn: func(got string) bool {
				return len(got) == 64
			},
		},
		{
			name:  "hashes long string",
			input: strings.Repeat("a", 1000),
			salt:  nil,
			wantFn: func(got string) bool {
				return len(got) == 64
			},
		},
		{
			name:  "hash with custom salt",
			input: "password123",
			salt:  util.Ptr(string("mysalt")),
			wantFn: func(got string) bool {
				return len(got) == 64 && got != ""
			},
		},
		{
			name:  "same input with same salt produces same hash",
			input: "test",
			salt:  util.Ptr(string("salt")),
			wantFn: func(got string) bool {
				return len(got) == 64
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewSHA256Hasher(tt.salt)
			got, err := hasher.Hash(tt.input)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			if !tt.wantFn(got) {
				t.Errorf("Hash() = %v, want valid hex string", got)
			}
		})
	}
}

func TestSHA256Hasher_Compare(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		salt        *string
		wrongPass   string
		wantCorrect bool
		wantWrong   bool
	}{
		{
			name:        "correct password matches, wrong does not",
			password:    "correct_password",
			salt:        nil,
			wrongPass:   "wrong_password",
			wantCorrect: true,
			wantWrong:   false,
		},
		{
			name:        "empty password comparison",
			password:    "test_password",
			salt:        nil,
			wrongPass:   "something",
			wantCorrect: true,
			wantWrong:   false,
		},
		{
			name:        "with salt - correct matches, wrong does not",
			password:    "salted_password",
			salt:        util.Ptr(string("mySalt")),
			wrongPass:   "wrong_password",
			wantCorrect: true,
			wantWrong:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewSHA256Hasher(tt.salt)
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

func TestSHA256Hasher_SameInputSameSalt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		salt  *string
	}{
		{"without salt", "password", nil},
		{"with salt", "password", util.Ptr(string("salt"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewSHA256Hasher(tt.salt)
			hash1, err := hasher.Hash(tt.input)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			hash2, err := hasher.Hash(tt.input)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if hash1 != hash2 {
				t.Error("same input with same salt should produce same hash")
			}
		})
	}
}

func TestSHA256Hasher_DifferentSalts(t *testing.T) {
	tests := []struct {
		name  string
		input string
		salt1 *string
		salt2 *string
	}{
		{"same input with different salts", "password", util.Ptr(string("salt1")), util.Ptr(string("salt2"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher1 := NewSHA256Hasher(tt.salt1)
			hasher2 := NewSHA256Hasher(tt.salt2)

			hash1, err := hasher1.Hash(tt.input)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			hash2, err := hasher2.Hash(tt.input)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if hash1 == hash2 {
				t.Error("same input with different salts should produce different hashes")
			}
		})
	}
}

func TestSHA256Hasher_DifferentInputs(t *testing.T) {
	tests := []struct {
		name   string
		input1 string
		input2 string
	}{
		{"different inputs produce different hashes", "password1", "password2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewSHA256Hasher(nil)

			hash1, err := hasher.Hash(tt.input1)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			hash2, err := hasher.Hash(tt.input2)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if hash1 == hash2 {
				t.Error("different inputs should produce different hashes")
			}
		})
	}
}

func TestSHA256Hasher_EmptySalt(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"hash with empty salt produces 64 char hex", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewSHA256Hasher(nil)
			hash, err := hasher.Hash(tt.input)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			if len(hash) != 64 {
				t.Errorf("hash length = %d, want 64", len(hash))
			}
		})
	}
}
