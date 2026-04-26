package observer

import (
	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

// NewPinObserver creates an observer for PIN service operations.
func NewPinObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.PinSignal] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.PinSignal) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
				attribute.String("user.uid", s.UserUID),
			}
			if s.Success != nil {
				attrs = append(attrs, attribute.Bool("success", *s.Success))
			}
			return attrs
		},
		func(s signal.PinSignal) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
				"user.uid":  s.UserUID,
			}
			if s.Success != nil {
				fields["success"] = *s.Success
			}
			return fields
		},
	)
}
