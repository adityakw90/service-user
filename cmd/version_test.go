package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVersionFlag(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantContains []string
		wantExitZero bool
	}{
		{
			name:         "version flag displays version and build time",
			args:         []string{"--version=true"},
			wantContains: []string{"service-user", "dev", "build:"},
			wantExitZero: true,
		},
		{
			name:         "version flag without value also works",
			args:         []string{"--version"},
			wantContains: []string{"service-user", "dev", "build:"},
			wantExitZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() {
				t.Skip("skipping integration test in short mode")
			}

			// Build test binary from repo root
			tempDir := t.TempDir()
			binaryPath := filepath.Join(tempDir, "test-service-user")
			repoRoot := ".."

			buildCtx, buildCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer buildCancel()
			buildCmd := exec.CommandContext(buildCtx, "go", "build", "-o", binaryPath, "./cmd")
			buildCmd.Dir = repoRoot
			var buildErr bytes.Buffer
			buildCmd.Stderr = &buildErr
			if err := buildCmd.Run(); err != nil {
				t.Fatalf("failed to build: %v\nstderr: %s", err, buildErr.String())
			}
			defer os.Remove(binaryPath)

			// Run with args
			runCtx, runCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer runCancel()
			cmd := exec.CommandContext(runCtx, binaryPath, tt.args...)
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			exitCode := 0
			if err := cmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}

			output := out.String()

			// Check exit code
			if tt.wantExitZero && exitCode != 0 {
				t.Errorf("expected exit code 0, got %d", exitCode)
			}

			// Check output contains expected strings
			for _, contains := range tt.wantContains {
				if !strings.Contains(output, contains) {
					t.Errorf("output %q does not contain %q", output, contains)
				}
			}
		})
	}
}
