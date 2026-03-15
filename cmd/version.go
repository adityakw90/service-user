package main

import (
	"fmt"
	"os"
)

var (
	// Version is set via ldflags during build (e.g., -X main.Version=1.0.0)
	Version = "dev"
	// BuildTime is set via ldflags during build (e.g., -X main.BuildTime=2024-03-12T10:30:00UTC)
	BuildTime = "unknown"
)

// handleVersionFlag checks if --version flag is set and prints version info.
// This must be called before config.Load() to handle version early.
// Note: os.Exit(0) will skip any defer statements in main().
func handleVersionFlag() {
	for _, arg := range os.Args {
		if arg == "--version=true" || arg == "--version" {
			fmt.Printf("service-user %s (build: %s)\n", Version, BuildTime)
			os.Exit(0)
		}
	}
}
