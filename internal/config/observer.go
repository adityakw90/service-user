package config

import "github.com/spf13/viper"

type ObserverConfig struct {
	Auth     bool `mapstructure:"auth"`
	User     bool `mapstructure:"user"`
	Device   bool `mapstructure:"device"`
	UserFile bool `mapstructure:"userfile"`
	Pin      bool `mapstructure:"pin"`
}

func defaultObserverConfig(key string, vConfig *viper.Viper) {
	vConfig.SetDefault(key+".auth", true)
	vConfig.SetDefault(key+".user", true)
	vConfig.SetDefault(key+".device", true)
	vConfig.SetDefault(key+".userfile", true)
	vConfig.SetDefault(key+".pin", true)
}
