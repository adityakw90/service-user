package observer

import (
	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

// NewUserObserver creates an observer for user service operations.
func NewUserObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.UserSignal] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.UserSignal) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
			}
			if s.UID != nil {
				attrs = append(attrs, attribute.String("user.uid", *s.UID))
			}
			if s.ActorUID != nil {
				attrs = append(attrs, attribute.String("actor.uid", *s.ActorUID))
			}
			if s.Username != nil {
				attrs = append(attrs, attribute.String("user.username", *s.Username))
			}
			if s.Email != nil {
				attrs = append(attrs, attribute.String("user.email", *s.Email))
			}
			if s.Status != nil {
				attrs = append(attrs, attribute.String("user.status", s.Status.String()))
			}
			if s.Active != nil {
				attrs = append(attrs, attribute.Bool("user.active", *s.Active))
			}
			if s.ChangesCount > 0 {
				attrs = append(attrs, attribute.Int("changes.count", s.ChangesCount))
			}
			return attrs
		},
		func(s signal.UserSignal) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
			}
			if s.UID != nil {
				fields["user.uid"] = *s.UID
			}
			if s.ActorUID != nil {
				fields["actor.uid"] = *s.ActorUID
			}
			if s.Username != nil {
				fields["user.username"] = *s.Username
			}
			if s.Email != nil {
				fields["user.email"] = *s.Email
			}
			if s.Status != nil {
				fields["user.status"] = s.Status.String()
			}
			if s.Active != nil {
				fields["user.active"] = *s.Active
			}
			if s.ChangesCount > 0 {
				fields["changes.count"] = s.ChangesCount
			}
			return fields
		},
	)
}
