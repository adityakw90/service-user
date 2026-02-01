package security

import (
	"testing"
)

func TestNewHasher(t *testing.T) {
	tests := []struct {
		name       string
		hasherType string
		params     map[string]any
		wantErr    bool
	}{
		{
			name:       "default argon2",
			hasherType: "",
			params:     nil,
			wantErr:    false,
		},
		{
			name:       "explicit argon2",
			hasherType: "argon2",
			params:     nil,
			wantErr:    false,
		},
		{
			name:       "bcrypt with cost",
			hasherType: "bcrypt",
			params:     map[string]any{"cost": 10},
			wantErr:    false,
		},
		{
			name:       "bcrypt with default cost",
			hasherType: "bcrypt",
			params:     nil,
			wantErr:    false,
		},
		{
			name:       "sha256 with salt",
			hasherType: "sha256",
			params:     map[string]any{"salt": "mysalt"},
			wantErr:    false,
		},
		{
			name:       "sha256 without salt",
			hasherType: "sha256",
			params:     nil,
			wantErr:    false,
		},
		{
			name:       "unknown type",
			hasherType: "md5",
			params:     nil,
			wantErr:    true,
		},
		{
			name:       "unsupported type",
			hasherType: "scrypt",
			params:     nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHasher(tt.hasherType, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHasher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Error("NewHasher() returned nil")
					return
				}
				hash, err := got.Hash("test_password")
				if err != nil {
					t.Errorf("hasher.Hash() error = %v", err)
				}
				if !got.Compare(hash, "test_password") {
					t.Error("hasher should verify correct password")
				}
			}
		})
	}
}

func TestNewHasher_Argon2WithParams(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]any
		wantErr bool
	}{
		{
			name: "custom memory",
			params: map[string]any{
				"Memory": uint32(64 * 1024),
			},
			wantErr: false,
		},
		{
			name: "custom iterations",
			params: map[string]any{
				"Iterations": uint32(2),
			},
			wantErr: false,
		},
		{
			name: "custom parallelism",
			params: map[string]any{
				"Parallelism": uint8(2),
			},
			wantErr: false,
		},
		{
			name: "custom salt length",
			params: map[string]any{
				"SaltLength": uint32(32),
			},
			wantErr: false,
		},
		{
			name: "custom key length",
			params: map[string]any{
				"KeyLength": uint32(64),
			},
			wantErr: false,
		},
		{
			name:    "empty params",
			params:  map[string]any{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHasher("argon2", tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHasher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				hash, err := got.Hash("test")
				if err != nil {
					t.Errorf("hasher.Hash() error = %v", err)
				}
				if !got.Compare(hash, "test") {
					t.Error("hasher should verify correct password")
				}
			}
		})
	}
}

func TestNewHasher_BCryptWithCost(t *testing.T) {
	tests := []struct {
		name    string
		cost    int
		wantErr bool
	}{
		{"min cost", 4, false},
		{"default cost", 10, false},
		{"high cost", 12, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHasher("bcrypt", map[string]any{"cost": tt.cost})
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHasher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				hash, err := got.Hash("test")
				if err != nil {
					t.Errorf("hasher.Hash() error = %v", err)
				}
				if !got.Compare(hash, "test") {
					t.Error("hasher should verify correct password")
				}
			}
		})
	}
}

func TestNewHasher_SHA256WithSalt(t *testing.T) {
	tests := []struct {
		name    string
		salt    string
		wantErr bool
	}{
		{"no salt", "", false},
		{"with salt", "mysalt", false},
		{"long salt", "a_very_long_salt_value_for_testing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHasher("sha256", map[string]any{"salt": tt.salt})
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHasher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				hash, err := got.Hash("test")
				if err != nil {
					t.Errorf("hasher.Hash() error = %v", err)
				}
				if !got.Compare(hash, "test") {
					t.Error("hasher should verify correct password")
				}
			}
		})
	}
}

func TestNewPINHasher(t *testing.T) {
	tests := []struct {
		name       string
		hasherType string
		wantErr    bool
	}{
		{
			name:       "default argon2",
			hasherType: "",
			wantErr:    false,
		},
		{
			name:       "explicit argon2",
			hasherType: "argon2",
			wantErr:    false,
		},
		{
			name:       "bcrypt",
			hasherType: "bcrypt",
			wantErr:    false,
		},
		{
			name:       "sha256",
			hasherType: "sha256",
			wantErr:    false,
		},
		{
			name:       "unknown type",
			hasherType: "md5",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPINHasher(tt.hasherType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPINHasher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Error("NewPINHasher() returned nil")
					return
				}
				pin := "123456"
				hash, err := got.Hash(pin)
				if err != nil {
					t.Errorf("hasher.Hash() error = %v", err)
				}
				if !got.Compare(hash, pin) {
					t.Error("hasher should verify correct PIN")
				}
				if got.Compare(hash, "000000") {
					t.Error("hasher should not verify wrong PIN")
				}
			}
		})
	}
}

func TestNewPINHasher_Argon2Optimized(t *testing.T) {
	tests := []struct {
		name string
		pin  string
	}{
		{"argon2 PIN hasher verifies PIN", "1234"},
		{"argon2 PIN hasher with longer PIN", "123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher, err := NewPINHasher("argon2")
			if err != nil {
				t.Fatalf("NewPINHasher() error = %v", err)
			}

			hash, err := hasher.Hash(tt.pin)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if !hasher.Compare(hash, tt.pin) {
				t.Error("PIN hasher should verify correct PIN")
			}
		})
	}
}

func TestNewPINHasher_BCryptHighCost(t *testing.T) {
	tests := []struct {
		name string
		pin  string
	}{
		{"bcrypt PIN hasher verifies PIN", "123456"},
		{"bcrypt PIN hasher with short PIN", "1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher, err := NewPINHasher("bcrypt")
			if err != nil {
				t.Fatalf("NewPINHasher() error = %v", err)
			}

			hash, err := hasher.Hash(tt.pin)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if !hasher.Compare(hash, tt.pin) {
				t.Error("PIN hasher should verify correct PIN")
			}
		})
	}
}

func TestNewHasher_FactoryConsistency(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"factory produces consistent hashers", "test_password_123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher1, err := NewHasher("argon2", map[string]any{})
			if err != nil {
				t.Fatalf("NewHasher() error = %v", err)
			}

			hasher2, err := NewHasher("argon2", map[string]any{})
			if err != nil {
				t.Fatalf("NewHasher() error = %v", err)
			}

			hash1, err := hasher1.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			hash2, err := hasher2.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if !hasher1.Compare(hash1, tt.password) {
				t.Error("hasher1 should verify password against hash1")
			}
			if !hasher1.Compare(hash2, tt.password) {
				t.Error("hasher1 should verify password against hash2")
			}
			if !hasher2.Compare(hash1, tt.password) {
				t.Error("hasher2 should verify password against hash1")
			}
			if !hasher2.Compare(hash2, tt.password) {
				t.Error("hasher2 should verify password against hash2")
			}
		})
	}
}
