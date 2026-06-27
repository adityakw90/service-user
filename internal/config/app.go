package config

import (
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AppConfig holds application-specific configuration.
type AppConfig struct {
	Name  string `mapstructure:"name"`
	Code  string `mapstructure:"code"`
	Env   string `mapstructure:"env"`
	Debug bool   `mapstructure:"debug"`
	IP    string `mapstructure:"ip"`
	Port  int    `mapstructure:"port"`
}

var flagOnce sync.Once

func defaultAppConfig(prefix string, v *viper.Viper) {
	v.SetDefault(prefix+".name", "Service User")
	v.SetDefault(prefix+".code", "SUS")
	v.SetDefault(prefix+".env", "development")
	v.SetDefault(prefix+".debug", false)
	v.SetDefault(prefix+".ip", "0.0.0.0")
	v.SetDefault(prefix+".port", 50051)
}

func InitFlags() {
	flagOnce.Do(func() {
		pflag.Bool("version", false, "Print version information and exit")
		pflag.String("config", "", "Path to configuration file")
		pflag.String("ip", "0.0.0.0", "Service listening IP")
		pflag.Int("port", 50051, "Service listening port")
		pflag.Parse()
	})
}
