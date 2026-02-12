package infra

import (
	gomon "github.com/adityakw90/go-monitoring"
	"go.opentelemetry.io/otel/trace"
)

// NoopLogger is a no-op logger implementation.
type NoopLogger struct{}

func (l *NoopLogger) SetLogLevel(level string)                            {}
func (l *NoopLogger) Debug(message string, fields map[string]interface{}) {}
func (l *NoopLogger) Info(message string, fields map[string]interface{})  {}
func (l *NoopLogger) Warn(message string, fields map[string]interface{})  {}
func (l *NoopLogger) Error(message string, fields map[string]interface{}) {}
func (l *NoopLogger) Fatal(message string, fields map[string]interface{}) {}
func (l *NoopLogger) WithSpanContext(span trace.SpanContext) gomon.Logger { return l }
func (l *NoopLogger) AddCallerSkipNum(skipNum int) gomon.Logger           { return l }
func (l *NoopLogger) Sync() error                                         { return nil }

func NewNoopLogger() gomon.Logger {
	return &NoopLogger{}
}
