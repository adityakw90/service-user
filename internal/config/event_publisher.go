package config

import (
	"time"

	"github.com/spf13/viper"
)

// EventPublisherConfig holds configuration for the async event publisher.
type EventPublisherConfig struct {
	Enabled bool `mapstructure:"enabled"`

	// Backends (can enable multiple)
	HTTP     PublisherHTTPConfig     `mapstructure:"http"`
	RabbitMQ PublisherRabbitMQConfig `mapstructure:"rabbitmq"`
}

type PublisherHTTPConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// RabbitMQPublisherConfig holds configuration for the RabbitMQ event publisher.
type PublisherRabbitMQConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	Exchange         string `mapstructure:"exchange"`
	RoutingKeyPrefix string `mapstructure:"routing_key_prefix"`

	// Publisher confirms (at-least-once delivery)
	ConfirmTimeout time.Duration `mapstructure:"confirm_timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
	RetryInterval  time.Duration `mapstructure:"retry_interval"`
}

// defaultEventPublisherConfig sets default values for event publisher configuration.
func defaultEventPublisherConfig(key string, v *viper.Viper) {
	v.SetDefault(key+".enabled", false)

	// http
	v.SetDefault(key+".http.enabled", false)
	v.SetDefault(key+".http.url", "http://localhost:8080")
	v.SetDefault(key+".http.timeout", 5*time.Second)

	// RabbitMQ defaults
	v.SetDefault(key+".rabbitmq.enabled", false)
	v.SetDefault(key+".rabbitmq.exchange", "user-service")
	v.SetDefault(key+".rabbitmq.routing_key_prefix", "user.service")
	v.SetDefault(key+".rabbitmq.confirm_timeout", 30*time.Second)
	v.SetDefault(key+".rabbitmq.max_retries", 5)
	v.SetDefault(key+".rabbitmq.retry_interval", 1*time.Second)
}
