package config

import "github.com/spf13/viper"

type InstanceConfig struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
}

func defaultInstanceConfig(prefix string, v *viper.Viper) {
	v.SetDefault(prefix+".name", "service-user")
	v.SetDefault(prefix+".host", "localhost")
}
