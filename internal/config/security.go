package config

import (
	"time"

	"github.com/spf13/viper"
)

// SecurityConfig holds configuration for security adapters.
type SecurityConfig struct {
	LoginTracker AttemptTrackerConfig `mapstructure:"login_tracker"` // Login attempt tracker settings
	PINTracker   AttemptTrackerConfig `mapstructure:"pin_tracker"`   // PIN attempt tracker settings
	RateLimiter  RateLimiterConfig    `mapstructure:"rate_limiter"`  // Rate limiting settings
}

// AttemptTrackerConfig holds configuration for an attempt tracker.
type AttemptTrackerConfig struct {
	Backend           string        `mapstructure:"backend"`             // which storage backend to use: "redis" or "memory"
	LockoutThreshold  int           `mapstructure:"lockout_threshold"`   // Number of failed attempts before lockout
	LockoutDuration   time.Duration `mapstructure:"lockout_duration"`    // How long to lock out the account
	LockoutCounterTTL time.Duration `mapstructure:"lockout_counter_ttl"` // How long to keep failed attempt counters
}

// RateLimiterConfig holds configuration for the rate limiter.
type RateLimiterConfig struct {
	Backend    string        `mapstructure:"backend"`     // which storage backend to use: "redis" or "memory"
	Limit      int           `mapstructure:"limit"`       // Maximum number of requests
	WindowSize time.Duration `mapstructure:"window_size"` // Time window for rate limiting
}

func defaultSecurityConfig(key string, vConfig *viper.Viper) {
	// Login tracker defaults
	vConfig.SetDefault(key+".login_tracker.backend", "redis")
	vConfig.SetDefault(key+".login_tracker.lockout_threshold", 5)
	vConfig.SetDefault(key+".login_tracker.lockout_duration", "15m")
	vConfig.SetDefault(key+".login_tracker.lockout_counter_ttl", "30m")

	// PIN tracker defaults (stricter than login)
	vConfig.SetDefault(key+".pin_tracker.backend", "redis")
	vConfig.SetDefault(key+".pin_tracker.lockout_threshold", 3)
	vConfig.SetDefault(key+".pin_tracker.lockout_duration", "15m")
	vConfig.SetDefault(key+".pin_tracker.lockout_counter_ttl", "30m")

	// Rate limiter defaults
	vConfig.SetDefault(key+".rate_limiter.backend", "redis")
	vConfig.SetDefault(key+".rate_limiter.limit", 100)
	vConfig.SetDefault(key+".rate_limiter.window_size", "1h")
}
