package testutil

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		setEnv       bool
		defaultValue string
		want         string
	}{
		{
			name:         "returns env value when set to non-empty string",
			key:          "TEST_VAR_NONEMPTY",
			envValue:     "test-value",
			setEnv:       true,
			defaultValue: "default",
			want:         "test-value",
		},
		{
			name:         "returns empty string when env var set to empty string (not default)",
			key:          "TEST_VAR_EMPTY",
			envValue:     "",
			setEnv:       true,
			defaultValue: "default",
			want:         "",
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_VAR_UNSET",
			envValue:     "",
			setEnv:       false,
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.Unsetenv(tt.key)

			// Set environment variable if required
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.defaultValue, got, tt.want)
			}
		})
	}
}
