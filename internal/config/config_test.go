package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoadWithCustomPath(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// Create test config file
	configContent := `
app:
  code: test-custom-config
  port: 9999
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set up flags to simulate --config flag
	// Reset pflag for clean test state
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.ContinueOnError)

	// Define flags that the production code expects
	pflag.String("config", "", "Path to configuration file")
	pflag.String("ip", "0.0.0.0", "Service listening IP")
	pflag.Int("port", 50051, "Service listening port")

	// Simulate command line args with --config flag
	os.Args = []string{"test", "--config", configPath}
	pflag.Parse()

	// Load config
	cfg, err := Load()
	require.NoError(t, err)

	// Verify config was loaded from custom path
	assert.Equal(t, "test-custom-config", cfg.App.Code, "App code should match custom config")
	assert.Equal(t, 9999, cfg.App.Port, "App port should match custom config")
}

func TestConfigLoadWithEtcPath(t *testing.T) {
	// Skip if not running as root (can't write to /etc)
	if os.Geteuid() != 0 {
		t.Skip("Skipping test: requires root privileges to write to /etc")
	}

	// Create /etc/service-user directory
	etcDir := "/etc/service-user"
	err := os.MkdirAll(etcDir, 0755)
	require.NoError(t, err)

	// Create test config file
	configPath := filepath.Join(etcDir, "config.yaml")
	configContent := `
app:
  code: test-etc-config
  port: 8888
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove(configPath)

	// Reset pflag for clean test state
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.ContinueOnError)

	// Define flags that the production code expects
	pflag.String("config", "", "Path to configuration file")
	pflag.String("ip", "0.0.0.0", "Service listening IP")
	pflag.Int("port", 50051, "Service listening port")

	// Simulate command line args without --config flag
	os.Args = []string{"test"}
	pflag.Parse()

	// Load config
	cfg, err := Load()
	require.NoError(t, err)

	// Verify config was loaded from /etc/service-user
	assert.Equal(t, "test-etc-config", cfg.App.Code, "App code should match /etc config")
	assert.Equal(t, 8888, cfg.App.Port, "App port should match /etc config")
}

func TestConfigLoadWithCurrentDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create config.yaml in current directory
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
app:
  code: test-current-dir-config
  port: 7777
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Change to test directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Reset pflag for clean test state
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.ContinueOnError)

	// Define flags that the production code expects
	pflag.String("config", "", "Path to configuration file")
	pflag.String("ip", "0.0.0.0", "Service listening IP")
	pflag.Int("port", 50051, "Service listening port")

	// Simulate command line args without --config flag
	os.Args = []string{"test"}
	pflag.Parse()

	// Load config
	cfg, err := Load()
	require.NoError(t, err)

	// Verify config was loaded from current directory
	assert.Equal(t, "test-current-dir-config", cfg.App.Code, "App code should match current dir config")
	assert.Equal(t, 7777, cfg.App.Port, "App port should match current dir config")
}

func TestConfigLoadPriority(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	etcDir := filepath.Join(tmpDir, "etc", "service-user")
	currentDir := tmpDir
	customDir := filepath.Join(tmpDir, "custom")

	err := os.MkdirAll(etcDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(customDir, 0755)
	require.NoError(t, err)

	// Create configs in all three locations
	// 1. /etc/service-user/config.yaml
	etcConfigPath := filepath.Join(etcDir, "config.yaml")
	etcConfigContent := `
app:
  code: etc-config
  port: 1111
`
	err = os.WriteFile(etcConfigPath, []byte(etcConfigContent), 0644)
	require.NoError(t, err)

	// 2. ./config.yaml (current directory)
	currentConfigPath := filepath.Join(currentDir, "config.yaml")
	currentConfigContent := `
app:
  code: current-dir-config
  port: 2222
`
	err = os.WriteFile(currentConfigPath, []byte(currentConfigContent), 0644)
	require.NoError(t, err)

	// 3. Custom path via --config flag
	customConfigPath := filepath.Join(customDir, "custom-config.yaml")
	customConfigContent := `
app:
  code: custom-path-config
  port: 3333
`
	err = os.WriteFile(customConfigPath, []byte(customConfigContent), 0644)
	require.NoError(t, err)

	// Change to test directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(currentDir)
	require.NoError(t, err)

	// Test 1: --config flag should have highest priority
	t.Run("--config flag priority", func(t *testing.T) {
		// Reset pflag for clean test state
		pflag.CommandLine = pflag.NewFlagSet("test", pflag.ContinueOnError)

		// Define flags that the production code expects
		pflag.String("config", "", "Path to configuration file")
		pflag.String("ip", "0.0.0.0", "Service listening IP")
		pflag.Int("port", 50051, "Service listening port")

		// Temporarily set AddConfigPath to use our test directories
		// We need to modify the Load function to accept config paths for testing
		// For now, we'll test by setting environment variable

		os.Args = []string{"test", "--config", customConfigPath}
		pflag.Parse()

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "custom-path-config", cfg.App.Code, "Custom config should have highest priority")
		assert.Equal(t, 3333, cfg.App.Port)
	})

	// Note: Testing /etc and current directory priority is complex
	// because the Load function uses hardcoded paths.
	// This would require refactoring Load to accept paths for testing.
}

func TestConfigLoadDefaults(t *testing.T) {
	// Create a temporary directory with no config file
	tmpDir := t.TempDir()

	// Change to test directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Reset pflag for clean test state
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.ContinueOnError)

	// Define flags that the production code expects
	pflag.String("config", "", "Path to configuration file")
	pflag.String("ip", "0.0.0.0", "Service listening IP")
	pflag.Int("port", 50051, "Service listening port")

	// Simulate command line args
	os.Args = []string{"test"}
	pflag.Parse()

	// Load config (should use defaults)
	cfg, err := Load()
	require.NoError(t, err)

	// Verify defaults
	assert.Equal(t, "SUS", cfg.App.Code, "Should use default app code")
	assert.Equal(t, 50051, cfg.App.Port, "Should use default port")
}
