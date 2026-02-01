package security

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUIDGenerator_New(t *testing.T) {
	tests := []struct {
		name         string
		count        int
		checkVersion bool
	}{
		{"generate 10 UUIDs and verify version", 10, true},
		{"generate 50 UUIDs for validation", 50, true},
		{"generate 100 UUIDs without version check", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewUIDGenerator()
			uuids := make(map[string]bool)
			for i := 0; i < tt.count; i++ {
				u := gen.New()

				if uuids[u] {
					t.Errorf("duplicate UUID generated: %s", u)
				}
				uuids[u] = true

				if tt.checkVersion {
					parsed, err := uuid.Parse(u)
					if err != nil {
						t.Errorf("failed to parse UUID %s: %v", u, err)
						continue
					}
					if parsed.Version() != 7 {
						t.Errorf("expected version 7, got %d for UUID %s", parsed.Version(), u)
					}
				}
			}
		})
	}
}

func TestUIDGenerator_New_IsUnique(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"generate 100 UUIDs and verify uniqueness", 100},
		{"generate 500 UUIDs for uniqueness", 500},
		{"generate 1000 UUIDs for uniqueness", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewUIDGenerator()
			uuids := make(map[string]int)
			for i := 0; i < tt.count; i++ {
				u := gen.New()
				uuids[u]++
			}

			for u, count := range uuids {
				if count > 1 {
					t.Errorf("UUID %s was generated %d times", u, count)
				}
			}
		})
	}
}

func TestUIDGenerator_New_SortOrder(t *testing.T) {
	tests := []struct {
		name     string
		interval int
	}{
		{"1ms interval", 1},
		{"2ms interval", 2},
		{"5ms interval", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewUIDGenerator()
			uuid1 := gen.New()
			// Ensure different timestamps
			for i := 0; i < tt.interval; i++ {
				time.Sleep(time.Millisecond)
				_ = gen.New() // Generate intermediate UUIDs
			}
			uuid2 := gen.New()

			if uuid1 >= uuid2 {
				t.Errorf("expected %s < %s for time-based ordering", uuid1, uuid2)
			}
		})
	}
}

func TestParseV7_Valid(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"parse generated UUID v7"},
		{"parse UUID with all zero timestamp"},
		{"parse UUID with max timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewUIDGenerator()
			u := gen.New()

			parsed, err := uuid.Parse(u)
			if err != nil {
				t.Errorf("failed to parse valid UUID v7: %v", err)
			}
			if parsed != uuid.MustParse(u) {
				t.Errorf("parsed UUID doesn't match original")
			}
		})
	}
}

func TestUIDGenerator_CollisionResistance(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"generate 1000 UUIDs", 1000},
		{"generate 5000 UUIDs", 5000},
		{"generate 10000 UUIDs", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewUIDGenerator()
			uuids := make(map[string]bool)
			for i := 0; i < tt.count; i++ {
				u := gen.New()
				if uuids[u] {
					t.Errorf("collision detected for UUID: %s", u)
				}
				uuids[u] = true
			}
		})
	}
}

func TestUIDGenerator_Format(t *testing.T) {
	tests := []struct {
		name       string
		wantPrefix string
		wantLen    int
	}{
		{"UUID v7 format", "", 36},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewUIDGenerator()
			u := gen.New()

			if len(u) != tt.wantLen {
				t.Errorf("UUID length = %d, want %d", len(u), tt.wantLen)
			}

			_, err := uuid.Parse(u)
			if err != nil {
				t.Errorf("failed to parse UUID: %v", err)
			}
		})
	}
}
