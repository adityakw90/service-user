package config

import "github.com/spf13/viper"

// MonitoringConfig holds monitoring-specific configuration.
type MonitoringConfig struct {
	ServiceName string                 `mapstructure:"service_name"`
	Environment string                 `mapstructure:"environment"`
	Logger      MonitoringLogConfig    `mapstructure:"logger"`
	Tracer      MonitoringTraceConfig  `mapstructure:"tracer"`
	Metric      MonitoringMetricConfig `mapstructure:"metric"`
}

type MonitoringLogConfig struct {
	Level string `mapstructure:"level"`
}

type MonitoringTraceConfig struct {
	Provider     string  `mapstructure:"provider"`      // "stdout", "jaeger", "otlp"
	ProviderHost string  `mapstructure:"provider_host"` // provider host
	ProviderPort int     `mapstructure:"provider_port"` // provider port
	SampleRatio  float64 `mapstructure:"sample_ratio"`  // provider port
}

type MonitoringMetricConfig struct {
	Provider     string `mapstructure:"provider"`      // "stdout", "jaeger", "otlp"
	ProviderHost string `mapstructure:"provider_host"` // provider host
	ProviderPort int    `mapstructure:"provider_port"` // provider port
}

func defaultMonitoringConfig(prefix string, v *viper.Viper) {
	v.SetDefault(prefix+".service_name", "service-user")
	v.SetDefault(prefix+".environment", "development")
	v.SetDefault(prefix+".logger.level", "info")
	v.SetDefault(prefix+".tracer.provider", "stdout")
	v.SetDefault(prefix+".tracer.provider_host", "localhost")
	v.SetDefault(prefix+".tracer.provider_port", 4317)
	v.SetDefault(prefix+".tracer.sample_ratio", 1.0)
	v.SetDefault(prefix+".metric.provider", "stdout")
	v.SetDefault(prefix+".metric.provider_host", "localhost")
	v.SetDefault(prefix+".metric.provider_port", 9411)
}
