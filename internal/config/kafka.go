package config

import (
	"reflect"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// KafkaConfig holds configuration for the Kafka event publisher.
type KafkaConfig struct {
	Brokers              []string                `mapstructure:"brokers"`
	MaxMessageBytes      int                     `mapstructure:"max_message_bytes"`
	TimeoutSeconds       int                     `mapstructure:"timeout_seconds"`
	Compression          sarama.CompressionCodec `mapstructure:"compression"`
	ReconnectMaxAttempts int                     `mapstructure:"reconnect_max_attempts"`
	ReconnectInterval    time.Duration           `mapstructure:"reconnect_interval_seconds"`
}

func defaultKafkaConfig(key string, vConfig *viper.Viper) {
	vConfig.SetDefault(key+".brokers", []string{"localhost:9092"})
	vConfig.SetDefault(key+".max_message_bytes", 1024*1024)
	vConfig.SetDefault(key+".timeout_seconds", 5)
	vConfig.SetDefault(key+".compression", "snappy")
	vConfig.SetDefault(key+".reconnect_max_attempts", 0)
	vConfig.SetDefault(key+".reconnect_interval_seconds", 1)
}

// kafkaCompressionHookFunc is a DecodeHookFunc that converts a string to a sarama.CompressionCodec.
func kafkaCompressionHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(sarama.CompressionCodec(0)) {
			return data, nil
		}

		// Convert data (string) to sarama.CompressionCodec
		var compression sarama.CompressionCodec
		err := compression.UnmarshalText([]byte(data.(string)))
		if err != nil {
			return nil, err
		}
		return compression, nil
	}
}
