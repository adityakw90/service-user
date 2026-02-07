package observer

import (
	"encoding/json"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

func NewAuthObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.AuthSignal] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.AuthSignal) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("identifier", s.Identifier),
				attribute.String("identifier_type", s.IdentifierType),
			}
			if s.DeviceFingerprint != nil {
				attrs = append(attrs, attribute.String("device.fingerprint", *s.DeviceFingerprint))
			}
			if s.DeviceIP != nil {
				attrs = append(attrs, attribute.String("device.ip", *s.DeviceIP))
			}
			if s.DeviceName != nil {
				attrs = append(attrs, attribute.String("device.name", *s.DeviceName))
			}
			if s.Extra != nil {
				jsonData, _ := json.Marshal(*s.Extra)
				attrs = append(attrs, attribute.String("extra", string(jsonData)))
			}
			if s.UID != nil {
				attrs = append(attrs, attribute.String("user.uid", *s.UID))
			}
			if s.Email != nil {
				attrs = append(attrs, attribute.String("user.email", *s.Email))
			}
			if s.Username != nil {
				attrs = append(attrs, attribute.String("user.username", *s.Username))
			}
			if s.Status != nil {
				attrs = append(attrs, attribute.String("user.status", string(*s.Status)))
			}
			if s.Active != nil {
				attrs = append(attrs, attribute.Bool("user.active", *s.Active))
			}
			if s.Deleted != nil {
				attrs = append(attrs, attribute.Bool("user.deleted", *s.Deleted))
			}
			return attrs
		},
		func(s signal.AuthSignal) map[string]any {
			fields := map[string]any{
				"identifier":      s.Identifier,
				"identifier_type": s.IdentifierType,
			}
			if s.DeviceFingerprint != nil {
				fields["device.fingerprint"] = *s.DeviceFingerprint
			}
			if s.DeviceIP != nil {
				fields["device.ip"] = *s.DeviceIP
			}
			if s.DeviceName != nil {
				fields["device.name"] = *s.DeviceName
			}
			if s.Extra != nil {
				jsonData, _ := json.Marshal(*s.Extra)
				fields["extra"] = string(jsonData)
			}
			if s.UID != nil {
				fields["user.uid"] = *s.UID
			}
			if s.Email != nil {
				fields["user.email"] = *s.Email
			}
			if s.Username != nil {
				fields["user.username"] = *s.Username
			}
			if s.Status != nil {
				fields["user.status"] = string(*s.Status)
			}
			if s.Active != nil {
				fields["user.active"] = *s.Active
			}
			if s.Deleted != nil {
				fields["user.deleted"] = *s.Deleted
			}
			return fields
		},
	)
}
