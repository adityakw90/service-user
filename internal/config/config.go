package config

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all configuration for the service.
type Config struct {
	App            AppConfig            `mapstructure:"app"`
	Instance       InstanceConfig       `mapstructure:"instance"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Redis          RedisConfig          `mapstructure:"redis"`
	Rabbit         RabbitConfig         `mapstructure:"rabbitmq"`
	Kafka          KafkaConfig          `mapstructure:"kafka"`
	Monitoring     MonitoringConfig     `mapstructure:"monitoring"`
	Observer       ObserverConfig       `mapstructure:"observer"`
	Security       SecurityConfig       `mapstructure:"security"`
	Jwt            JWTConfig            `mapstructure:"jwt"`
	PasswordHasher HasherConfig         `mapstructure:"password_hasher"`
	PINHasher      HasherConfig         `mapstructure:"pin_hasher"`
	EventPublisher EventPublisherConfig `mapstructure:"event_publisher"`
	OAuth          OauthConfig          `mapstructure:"oauth"`
}

// Load reads configuration from environment variables using Viper.
func Load() (*Config, error) {
	// Use pflag (a better command-line flag package compatible with flag)
	pflag.String("config", "", "Path to configuration file")
	pflag.String("ip", "0.0.0.0", "Service listening IP")
	pflag.Int("port", 50051, "Service listening port")
	pflag.Parse()

	// prepare decoder
	// Create a DecoderConfigOption with the custom hook
	decodeOption := viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			dc.DecodeHook, // Keep existing hooks
			kafkaCompressionHookFunc(),
		)
	})

	// initialize
	vConfig := viper.New()

	// Manually bind flags to specific keys in Viper
	vConfig.BindPFlag("app.ip", pflag.Lookup("ip"))
	vConfig.BindPFlag("app.port", pflag.Lookup("port"))

	// Config file search order:
	// 1. --config flag (highest priority)
	// 2. /etc/service-user/config.yaml (system location)
	// 3. ./config.yaml (current directory, fallback)
	configFlag := pflag.Lookup("config").Value.String()
	if configFlag != "" {
		vConfig.SetConfigFile(configFlag)
	} else {
		vConfig.SetConfigName("config")
		vConfig.SetConfigType("yaml")
		// NOTE: Viper searches paths in REVERSE order of addition.
		// To search /etc/service-user first, we add it LAST.
		vConfig.AddConfigPath(".")                 // Current directory (fallback)
		vConfig.AddConfigPath("/etc/service-user") // System location (primary)
	}

	// default config
	defaultAppConfig("app", vConfig)
	defaultInstanceConfig("instance", vConfig)
	defaultDatabaseConfig("database", vConfig)
	defaultRedisConfig("redis", vConfig)
	defaultRabbitConfig("rabbitmq", vConfig)
	defaultKafkaConfig("kafka", vConfig)
	defaultMonitoringConfig("monitoring", vConfig)
	defaultObserverConfig("observer", vConfig)
	defaultJWTConfig("jwt", vConfig)
	defaultHasherConfig("password_hasher", vConfig)
	defaultHasherConfig("pin_hasher", vConfig)
	defaultSecurityConfig("security", vConfig)
	defaultEventPublisherConfig("event_publisher", vConfig)
	defaultOauthConfig("oauth", vConfig)

	// test
	// vConfig.SafeWriteConfig() // Removed - not appropriate for binary deployments

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
	if err := vConfig.Unmarshal(&cfg, decodeOption); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
