package config

import (
	"time"

	"github.com/spf13/viper"
)

// EventPublisherConfig holds configuration for the async event publisher.
type EventPublisherConfig struct {
	// Async settings
	Enabled     bool `mapstructure:"enabled"`
	WorkerCount int  `mapstructure:"worker_count"`
	QueueSize   int  `mapstructure:"queue_size"`

	// Batching settings
	BatchSize    int           `mapstructure:"batch_size"`
	BatchTimeout time.Duration `mapstructure:"batch_timeout"`

	// Retry settings for async publisher (for failed publishes)
	MaxRetries    int           `mapstructure:"max_retries"`
	RetryInterval time.Duration `mapstructure:"retry_interval"`

	// Backends (can enable multiple)
	HTTP     PublisherHTTPConfig     `mapstructure:"http"`
	Redis    PublisherRedisConfig    `mapstructure:"redis"`
	RabbitMQ PublisherRabbitMQConfig `mapstructure:"rabbitmq"`
	Kafka    PublisherKafkaConfig    `mapstructure:"kafka"`
}

type PublisherHTTPConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// PublisherRedisConfig holds configuration for the Redis Stream event publisher.
type PublisherRedisConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Name    string `mapstructure:"name"`
	MaxLen  int64  `mapstructure:"max_len"`
}

// PublisherKafkaConfig holds configuration for the Kafka event publisher.
type PublisherKafkaConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Topic   string `mapstructure:"topic"`
}

// RabbitMQPublisherConfig holds configuration for the RabbitMQ event publisher.
type PublisherRabbitMQConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	Exchange         string        `mapstructure:"exchange"`
	ExchangeType     string        `mapstructure:"exchange_type"`
	RoutingKeyPrefix string        `mapstructure:"routing_key_prefix"`
	Durable          bool          `mapstructure:"durable"`
	// Publisher confirms (at-least-once delivery)
	ConfirmTimeout time.Duration `mapstructure:"confirm_timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
	RetryInterval  time.Duration `mapstructure:"retry_interval"`
	// Queue declaration (stores messages when no consumers are running)
	QueueName       string `mapstructure:"queue_name"`
	QueueDurable    bool   `mapstructure:"queue_durable"`
	QueueAutoDelete bool   `mapstructure:"queue_auto_delete"`
	QueueExclusive  bool   `mapstructure:"queue_exclusive"`
	QueueEnabled    bool   `mapstructure:"queue_enabled"`
}

// defaultEventPublisherConfig sets default values for event publisher configuration.
func defaultEventPublisherConfig(key string, v *viper.Viper) {
	v.SetDefault(key+".enabled", true)
	v.SetDefault(key+".worker_count", 2)
	v.SetDefault(key+".queue_size", 1000)
	v.SetDefault(key+".batch_size", 50)
	v.SetDefault(key+".batch_timeout", 5*time.Second)

	// Async publisher retry defaults
	v.SetDefault(key+".max_retries", 3)
	v.SetDefault(key+".retry_interval", 2*time.Second)

	// http
	v.SetDefault(key+".http.enabled", false)
	v.SetDefault(key+".http.url", "http://localhost:8080")
	v.SetDefault(key+".http.timeout", 5*time.Second)

	// redis stream
	v.SetDefault(key+".redis.enabled", true)
	v.SetDefault(key+".redis.name", "service-user.events")
	v.SetDefault(key+".redis.max_len", 100000)

	// Kafka defaults
	v.SetDefault(key+".kafka.enabled", false)
	v.SetDefault(key+".kafka.topic", "user-service-events")

	// RabbitMQ defaults
	v.SetDefault(key+".rabbitmq.enabled", false)
	v.SetDefault(key+".rabbitmq.exchange", "user-service")
	v.SetDefault(key+".rabbitmq.exchange_type", "topic")
	v.SetDefault(key+".rabbitmq.routing_key_prefix", "user.service.")
	v.SetDefault(key+".rabbitmq.durable", true)
	v.SetDefault(key+".rabbitmq.confirm_timeout", 30*time.Second)
	v.SetDefault(key+".rabbitmq.max_retries", 5)
	v.SetDefault(key+".rabbitmq.retry_interval", 1*time.Second)
	// Queue defaults (enabled by default for at-least-once delivery)
	v.SetDefault(key+".rabbitmq.queue_enabled", true)
	v.SetDefault(key+".rabbitmq.queue_name", "user-service.queue")
	v.SetDefault(key+".rabbitmq.queue_durable", true)
	v.SetDefault(key+".rabbitmq.queue_auto_delete", false)
	v.SetDefault(key+".rabbitmq.queue_exclusive", false)
}
