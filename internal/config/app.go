package config

import "github.com/spf13/viper"

// AppConfig holds application-specific configuration.
type AppConfig struct {
	Name  string `mapstructure:"name"`
	Code  string `mapstructure:"code"`
	Env   string `mapstructure:"env"`
	Debug bool   `mapstructure:"debug"`
	IP    string `mapstructure:"ip"`
	Port  int    `mapstructure:"port"`
}

func defaultAppConfig(prefix string, v *viper.Viper) {
	v.SetDefault(prefix+".name", "Service User")
	v.SetDefault(prefix+".code", "SUS")
	v.SetDefault(prefix+".env", "development")
	v.SetDefault(prefix+".debug", false)
	v.SetDefault(prefix+".ip", "0.0.0.0")
	v.SetDefault(prefix+".port", 50051)
}
