package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all configuration for the service.
type Config struct {
	App            AppConfig            `mapstructure:"app"`
	Instance       InstanceConfig       `mapstructure:"instance"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Redis          RedisConfig          `mapstructure:"redis"`
	Monitoring     MonitoringConfig     `mapstructure:"monitoring"`
	Observer       ObserverConfig       `mapstructure:"observer"`
	Security       SecurityConfig       `mapstructure:"security"`
	Jwt            JWTConfig            `mapstructure:"jwt"`
	PasswordHasher HasherConfig         `mapstructure:"password_hasher"`
	PINHasher      HasherConfig         `mapstructure:"pin_hasher"`
	EventPublisher EventPublisherConfig `mapstructure:"event_publisher"`

	// OAuth settings
	OAuthGoogleClientID     string `mapstructure:"OAUTH_GOOGLE_CLIENT_ID"`
	OAuthGoogleClientSecret string `mapstructure:"OAUTH_GOOGLE_CLIENT_SECRET"`
	OAuthRedirectURI        string `mapstructure:"OAUTH_REDIRECT_URI"`
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
	defaultInstanceConfig("instance", vConfig)
	defaultDatabaseConfig("database", vConfig)
	defaultRedisConfig("redis", vConfig)
	defaultMonitoringConfig("monitoring", vConfig)
	defaultObserverConfig("observer", vConfig)
	defaultJWTConfig("jwt", vConfig)
	defaultHasherConfig("password_hasher", vConfig)
	defaultHasherConfig("pin_hasher", vConfig)
	defaultSecurityConfig("security", vConfig)
	defaultEventPublisherConfig("event_publisher", vConfig)

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
