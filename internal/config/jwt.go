package config

import (
	"time"

	"github.com/spf13/viper"
)

type JWTConfig struct {
	SecretKey     string        `mapstructure:"secret_key"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
}

func defaultJWTConfig(prefix string, v *viper.Viper) {
	v.SetDefault(prefix+".secret_key", "your-256-bit-secret-key-for-jwt-signing")
	v.SetDefault(prefix+".access_expiry", 15*time.Minute)
	v.SetDefault(prefix+".refresh_expiry", 168*time.Hour) // 7 days
}
