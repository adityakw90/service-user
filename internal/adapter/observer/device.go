package observer

import (
	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

// NewDeviceObserver creates an observer for device service operations.
func NewDeviceObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.DeviceSignal] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.DeviceSignal) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
			}
			if s.UID != nil {
				attrs = append(attrs, attribute.String("device.uid", *s.UID))
			}
			if s.UserUID != nil {
				attrs = append(attrs, attribute.String("user.uid", *s.UserUID))
			}
			if s.DeviceName != nil {
				attrs = append(attrs, attribute.String("device.name", *s.DeviceName))
			}
			if s.Fingerprint != nil {
				attrs = append(attrs, attribute.String("device.fingerprint", *s.Fingerprint))
			}
			if s.IPAddress != nil {
				attrs = append(attrs, attribute.String("device.ip", *s.IPAddress))
			}
			return attrs
		},
		func(s signal.DeviceSignal) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
			}
			if s.UID != nil {
				fields["device.uid"] = *s.UID
			}
			if s.UserUID != nil {
				fields["user.uid"] = *s.UserUID
			}
			if s.DeviceName != nil {
				fields["device.name"] = *s.DeviceName
			}
			if s.Fingerprint != nil {
				fields["device.fingerprint"] = *s.Fingerprint
			}
			if s.IPAddress != nil {
				fields["device.ip"] = *s.IPAddress
			}
			return fields
		},
	)
}
