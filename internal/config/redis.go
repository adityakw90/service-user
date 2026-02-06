package config

import (
	"time"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	DB                int           `mapstructure:"db"`
	PoolSize          int           `mapstructure:"pool_size"`
	PoolTimeout       time.Duration `mapstructure:"pool_timeout"`
	ConnectionIdleMin int           `mapstructure:"connection_idle_min"`
}

func defaultRedisConfig(key string, vConfig *viper.Viper) {
	vConfig.SetDefault(key+".host", "localhost")
	vConfig.SetDefault(key+".port", 6379)
	vConfig.SetDefault(key+".user", "")
	vConfig.SetDefault(key+".password", "")
	vConfig.SetDefault(key+".db", 0)
	vConfig.SetDefault(key+".pool_size", 10)
	vConfig.SetDefault(key+".pool_timeout", "5s")
	vConfig.SetDefault(key+".connection_idle_min", 10)
}
