package config

import (
	"time"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host                  string        `mapstructure:"host"`
	Port                  int           `mapstructure:"port"`
	User                  string        `mapstructure:"user"`
	Password              string        `mapstructure:"password"`
	Name                  string        `mapstructure:"name"`
	SslMode               string        `mapstructure:"sslmode"`
	Timezone              string        `mapstructure:"timezone"`
	MinConns              int32         `mapstructure:"min_conns"`
	MinIdleConns          int32         `mapstructure:"min_idle_conns"`
	MaxConns              int32         `mapstructure:"max_conns"`
	MaxConnIdleTime       time.Duration `mapstructure:"max_conn_idle_time"`
	MaxConnLifetime       time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnLifetimeJitter time.Duration `mapstructure:"max_conn_lifetime_jitter"`
	HealthCheckPeriod     time.Duration `mapstructure:"health_check_period"`
}

func defaultDatabaseConfig(key string, vConfig *viper.Viper) {
	vConfig.SetDefault(key+".host", "localhost")
	vConfig.SetDefault(key+".port", 5432)
	vConfig.SetDefault(key+".user", "user")
	vConfig.SetDefault(key+".password", "password")
	vConfig.SetDefault(key+".name", "service_user")
	vConfig.SetDefault(key+".sslmode", "disable")
	vConfig.SetDefault(key+".timezone", "UTC")
	vConfig.SetDefault(key+".min_conns", int32(1))
	vConfig.SetDefault(key+".min_idle_conns", int32(0))
	vConfig.SetDefault(key+".max_conns", int32(4))
	vConfig.SetDefault(key+".max_conn_idle_time", "5m")
	vConfig.SetDefault(key+".max_conn_lifetime", "1h")
	vConfig.SetDefault(key+".max_conn_lifetime_jitter", "1m")
	vConfig.SetDefault(key+".health_check_period", "1m")
}
