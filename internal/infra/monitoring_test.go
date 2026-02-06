package infra

import (
	"testing"
)

func TestInfra_NewMonitoring(t *testing.T) {
	cfg := &MonitoringConfig{
		ServiceName:        "test-service",
		Environment:        "development",
		InstanceName:       "test-instance",
		InstanceHost:       "test-host",
		LoggerLevel:        "debug",
		TracerProvider:     "otlp",
		TracerProviderHost: "localhost",
		TracerProviderPort: 6831,
		TracerSampleRatio:  1.0,
		MetricProvider:     "otlp",
		MetricProviderHost: "localhost",
		MetricProviderPort: 9090,
	}

	monitoring, err := NewMonitoring(cfg)
	if err != nil {
		t.Errorf("NewMonitoring() error = %v", err)
	}

	if monitoring == nil {
		t.Errorf("NewMonitoring() returned nil")
	}
}
