package infra

import (
	"time"

	gomon "github.com/adityakw90/go-monitoring"
)

type MonitoringConfig struct {
	ServiceName         string
	Environment         string
	InstanceName        string
	InstanceHost        string
	LoggerLevel         string
	LoggerCallerSkipNum int
	TracerProvider      string
	TracerProviderHost  string
	TracerProviderPort  int
	TracerSampleRatio   float64
	TracerInsecure      bool
	MetricProvider      string
	MetricProviderHost  string
	MetricProviderPort  int
	MetricInsecure      bool
}

// ConvertConfigToOptions converts the old config structure to new go-monitoring options
func ConvertConfigToOptions(cfg *MonitoringConfig) []gomon.Option {
	opts := []gomon.Option{
		gomon.WithServiceName(cfg.ServiceName),
	}

	if cfg.Environment != "" {
		opts = append(opts, gomon.WithEnvironment(cfg.Environment))
	}

	if cfg.InstanceName != "" || cfg.InstanceHost != "" {
		opts = append(opts, gomon.WithInstance(cfg.InstanceName, cfg.InstanceHost))
	}

	if cfg.LoggerLevel != "" {
		opts = append(opts, gomon.WithLoggerLevel(cfg.LoggerLevel))
	}

	if cfg.LoggerCallerSkipNum > 0 {
		opts = append(opts, gomon.WithLoggerCallerSkipNum(cfg.LoggerCallerSkipNum))
	}

	if cfg.TracerProvider != "" {
		opts = append(opts, gomon.WithTracerProvider(
			cfg.TracerProvider,
			cfg.TracerProviderHost,
			cfg.TracerProviderPort,
		))
		if cfg.TracerSampleRatio > 0 {
			opts = append(opts, gomon.WithTracerSampleRatio(cfg.TracerSampleRatio))
		}
		if cfg.TracerInsecure {
			opts = append(opts, gomon.WithTracerInsecure(cfg.TracerInsecure))
		}
	}

	if cfg.MetricProvider != "" {
		opts = append(opts, gomon.WithMetricProvider(
			cfg.MetricProvider,
			cfg.MetricProviderHost,
			cfg.MetricProviderPort,
		))
		if cfg.MetricInsecure {
			opts = append(opts, gomon.WithMetricInsecure(cfg.MetricInsecure))
		}
		// Default metric interval if not specified
		opts = append(opts, gomon.WithMetricInterval(60*time.Second))
	}

	return opts
}

// NewMonitoring initializes the new go-monitoring library from config
func NewMonitoring(cfg *MonitoringConfig) (*gomon.Monitoring, error) {
	opts := ConvertConfigToOptions(cfg)
	return gomon.NewMonitoring(opts...)
}

// NewLogger creates a logger using the new go-monitoring library
func NewLogger() gomon.Logger {
	logger, err := gomon.NewLogger()
	if err != nil {
		// Fallback to a basic logger if initialization fails
		// This should rarely happen
		panic(err)
	}
	return logger
}
