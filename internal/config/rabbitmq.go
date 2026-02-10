package config

import (
	"time"

	"github.com/spf13/viper"
)

type RabbitConfig struct {
	Host                 string        `mapstructure:"host"`
	Port                 int           `mapstructure:"port"`
	User                 string        `mapstructure:"user"`
	Password             string        `mapstructure:"password"`
	Vhost                string        `mapstructure:"vhost"`
	ReconnectInterval    time.Duration `mapstructure:"reconnect_interval"`
	ReconnectMaxAttempts int           `mapstructure:"reconnect_max_attempts"`
}

func defaultRabbitConfig(key string, vConfig *viper.Viper) {
	vConfig.SetDefault(key+".host", "localhost")
	vConfig.SetDefault(key+".port", 5672)
	vConfig.SetDefault(key+".user", "guest")
	vConfig.SetDefault(key+".password", "guest")
	vConfig.SetDefault(key+".vhost", "/")
	vConfig.SetDefault(key+".reconnect_interval", 1*time.Second)
	vConfig.SetDefault(key+".reconnect_max_attempts", 0)
}
