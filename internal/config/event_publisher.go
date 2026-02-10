package config

import (
	"time"

	"github.com/spf13/viper"
)

// EventPublisherConfig holds configuration for the async event publisher.
type EventPublisherConfig struct {
	// Async settings
	Enabled    bool          `mapstructure:"enabled"`
	WorkerCount int          `mapstructure:"worker_count"`
	QueueSize   int          `mapstructure:"queue_size"`

	// Batching settings
	BatchSize    int           `mapstructure:"batch_size"`
	BatchTimeout time.Duration `mapstructure:"batch_timeout_seconds"`

	// Backends (can enable multiple)
	RedisPubSub  bool   `mapstructure:"redis_pubsub"`
	RedisChannel string `mapstructure:"redis_channel"`

	// Kafka
	Kafka KafkaConfig `mapstructure:"kafka"`

	// RabbitMQ
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`

	// HTTP
	HTTPEndpoint string `mapstructure:"http_endpoint"`
}

// KafkaConfig holds configuration for the Kafka event publisher.
type KafkaConfig struct {
	Enabled         bool     `mapstructure:"enabled"`
	Brokers         []string `mapstructure:"brokers"`
	Topic           string   `mapstructure:"topic"`
	MaxMessageBytes int      `mapstructure:"max_message_bytes"`
	TimeoutSeconds  int      `mapstructure:"timeout_seconds"`
	Compression     string   `mapstructure:"compression"`
}

// RabbitMQConfig holds configuration for the RabbitMQ event publisher.
type RabbitMQConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	URL              string `mapstructure:"url"`
	Exchange         string `mapstructure:"exchange"`
	ExchangeType     string `mapstructure:"exchange_type"`
	RoutingKeyPrefix string `mapstructure:"routing_key_prefix"`
	Durable          bool   `mapstructure:"durable"`

	// Reconnection settings (optional, defaults provided)
	ReconnectInterval    int `mapstructure:"reconnect_interval_seconds"`
	MaxReconnectAttempts int `mapstructure:"max_reconnect_attempts"`
	ReconnectDelay       int `mapstructure:"reconnect_delay_seconds"`
}

// defaultEventPublisherConfig sets default values for event publisher configuration.
func defaultEventPublisherConfig(key string, v *viper.Viper) {
	v.SetDefault(key+".enabled", true)
	v.SetDefault(key+".worker_count", 2)
	v.SetDefault(key+".queue_size", 1000)
	v.SetDefault(key+".batch_size", 50)
	v.SetDefault(key+".batch_timeout_seconds", 5)
	v.SetDefault(key+".redis_pubsub", true)
	v.SetDefault(key+".redis_channel", "user-service:events")

	// Kafka defaults
	v.SetDefault(key+".kafka.enabled", false)
	v.SetDefault(key+".kafka.brokers", []string{"localhost:9092"})
	v.SetDefault(key+".kafka.topic", "user-service-events")
	v.SetDefault(key+".kafka.max_message_bytes", 1048576)
	v.SetDefault(key+".kafka.timeout_seconds", 10)
	v.SetDefault(key+".kafka.compression", "snappy")

	// RabbitMQ defaults
	v.SetDefault(key+".rabbitmq.enabled", false)
	v.SetDefault(key+".rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	v.SetDefault(key+".rabbitmq.exchange", "user-service")
	v.SetDefault(key+".rabbitmq.exchange_type", "topic")
	v.SetDefault(key+".rabbitmq.routing_key_prefix", "user.service.")
	v.SetDefault(key+".rabbitmq.durable", true)
	v.SetDefault(key+".rabbitmq.reconnect_interval_seconds", 5)
	v.SetDefault(key+".rabbitmq.max_reconnect_attempts", 0) // 0 means infinite retries
	v.SetDefault(key+".rabbitmq.reconnect_delay_seconds", 1)

	// HTTP defaults
	v.SetDefault(key+".http_endpoint", "")
}
