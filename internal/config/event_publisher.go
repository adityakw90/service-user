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
	BatchTimeout time.Duration `mapstructure:"batch_timeout_seconds"`

	// Backends (can enable multiple)
	RedisPubSub  bool   `mapstructure:"redis_pubsub"`
	RedisChannel string `mapstructure:"redis_channel"`

	// Kafka
	Kafka PublisherKafkaConfig `mapstructure:"kafka"`

	// RabbitMQ
	RabbitMQ PublisherRabbitMQConfig `mapstructure:"rabbitmq"`

	// HTTP
	HTTPEndpoint string        `mapstructure:"http_endpoint"`
	HTTPTimeout  time.Duration `mapstructure:"http_timeout_seconds"`
}

// KafkaConfig holds configuration for the Kafka event publisher.
type PublisherKafkaConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Topic   string `mapstructure:"topic"`
}

// RabbitMQPublisherConfig holds configuration for the RabbitMQ event publisher.
type PublisherRabbitMQConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	Exchange         string `mapstructure:"exchange"`
	ExchangeType     string `mapstructure:"exchange_type"`
	RoutingKeyPrefix string `mapstructure:"routing_key_prefix"`
	Durable          bool   `mapstructure:"durable"`
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
	v.SetDefault(key+".kafka.topic", "user-service-events")

	// RabbitMQ defaults
	v.SetDefault(key+".rabbitmq.enabled", false)
	v.SetDefault(key+".rabbitmq.exchange", "user-service")
	v.SetDefault(key+".rabbitmq.exchange_type", "topic")
	v.SetDefault(key+".rabbitmq.routing_key_prefix", "user.service.")
	v.SetDefault(key+".rabbitmq.durable", true)

	// HTTP defaults
	v.SetDefault(key+".http_endpoint", "")
	v.SetDefault(key+".http_timeout_seconds", 5)
}
