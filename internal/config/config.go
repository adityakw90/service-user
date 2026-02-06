package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all configuration for the service.
type Config struct {
	App            AppConfig      `mapstructure:"app"`
	Database       DatabaseConfig `mapstructure:"database"`
	Redis          RedisConfig    `mapstructure:"redis"`
	Jwt            JWTConfig      `mapstructure:"jwt"`
	PasswordHasher HasherConfig   `mapstructure:"password_hasher"`
	PINHasher      HasherConfig   `mapstructure:"pin_hasher"`

	// Lockout settings
	LockoutMaxAttempts int           `mapstructure:"LOCKOUT_MAX_ATTEMPTS"`
	LockoutDuration    time.Duration `mapstructure:"LOCKOUT_DURATION_MINUTES"`

	// Rate limiting settings
	RateLimitLimit  int           `mapstructure:"RATE_LIMIT_LIMIT"`
	RateLimitWindow time.Duration `mapstructure:"RATE_LIMIT_WINDOW_SECONDS"`

	// OAuth settings
	OAuthGoogleClientID     string `mapstructure:"OAUTH_GOOGLE_CLIENT_ID"`
	OAuthGoogleClientSecret string `mapstructure:"OAUTH_GOOGLE_CLIENT_SECRET"`
	OAuthRedirectURI        string `mapstructure:"OAUTH_REDIRECT_URI"`

	// Event publisher settings
	EventPublisherEndpoint string        `mapstructure:"EVENT_PUBLISHER_ENDPOINT"`
	EventPublisherSource   string        `mapstructure:"EVENT_PUBLISHER_SOURCE"`
	EventPublisherTimeout  time.Duration `mapstructure:"EVENT_PUBLISHER_TIMEOUT_SECONDS"`

	// Monitoring settings
	Monitoring MonitoringConfig `mapstructure:",squash"`
}

type InstanceConfig struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
}

type MonitoringConfig struct {
	ServiceName string                 `mapstructure:"service_name"`
	Environment string                 `mapstructure:"environment"`
	Logger      MonitoringLogConfig    `mapstructure:"logger"`
	Tracer      MonitoringTraceConfig  `mapstructure:"tracer"`
	Metric      MonitoringMetricConfig `mapstructure:"metric"`
	Instance    InstanceConfig
}

type MonitoringLogConfig struct {
	Level string `mapstructure:"level"`
}

type MonitoringTraceConfig struct {
	Provider     string  `mapstructure:"provider"`      // "stdout", "jaeger", "otlp"
	ProviderHost string  `mapstructure:"provider_host"` // provider host
	ProviderPort int     `mapstructure:"provider_port"` // provider port
	SampleRatio  float64 `mapstructure:"sample_ratio"`  // provider port
}

type MonitoringMetricConfig struct {
	Provider     string `mapstructure:"provider"`      // "stdout", "jaeger", "otlp"
	ProviderHost string `mapstructure:"provider_host"` // provider host
	ProviderPort int    `mapstructure:"provider_port"` // provider port
}

// Load reads configuration from environment variables using Viper.
func Load() (*Config, error) {
	// Use pflag (a better command-line flag package compatible with flag)
	pflag.String("ip", "0.0.0.0", "Service listening IP")
	pflag.Int("port", 50051, "Service listening port")
	pflag.Parse()

	// initialize
	vConfig := viper.New()

	// Manually bind flags to specific keys in Viper
	vConfig.BindPFlag("app.ip", pflag.Lookup("ip"))
	vConfig.BindPFlag("app.port", pflag.Lookup("port"))

	// set config file
	vConfig.SetConfigName("config")
	vConfig.SetConfigType("yaml")
	vConfig.AddConfigPath(".")

	// default config
	defaultAppConfig("app", vConfig)
	defaultDatabaseConfig("database", vConfig)
	defaultRedisConfig("redis", vConfig)
	defaultJWTConfig("jwt", vConfig)
	defaultHasherConfig("password_hasher", vConfig)
	defaultHasherConfig("pin_hasher", vConfig)

	// Set defaults
	vConfig.SetDefault("LOCKOUT_MAX_ATTEMPTS", 5)
	vConfig.SetDefault("LOCKOUT_DURATION_MINUTES", 15)
	vConfig.SetDefault("RATE_LIMIT_LIMIT", 10)
	vConfig.SetDefault("RATE_LIMIT_WINDOW_SECONDS", 60)
	vConfig.SetDefault("EVENT_PUBLISHER_SOURCE", "service-user")
	vConfig.SetDefault("EVENT_PUBLISHER_TIMEOUT_SECONDS", 5)
	vConfig.SetDefault("MONITORING_SERVICE_NAME", "service-user")
	vConfig.SetDefault("MONITORING_ENVIRONMENT", "development")

	// test
	vConfig.SafeWriteConfig()

	// Enable environment variable override
	vConfig.AutomaticEnv()

	// Use a replacer to convert dots (.) in config keys to underscores (_) in env variables
	vConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read in the config file
	if err := vConfig.ReadInConfig(); err != nil {
		// Ignore error if config file is not found
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		} else {
			return nil, err
		}
	}

	var cfg Config
	if err := vConfig.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
