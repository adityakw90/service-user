package observer

import (
	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

// NewUserFileObserver creates an observer for user file service operations.
func NewUserFileObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.UserFileSignal] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.UserFileSignal) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
			}
			if s.UID != nil {
				attrs = append(attrs, attribute.String("file.uid", *s.UID))
			}
			if s.UserUID != nil {
				attrs = append(attrs, attribute.String("user.uid", *s.UserUID))
			}
			if s.FileName != nil {
				attrs = append(attrs, attribute.String("file.name", *s.FileName))
			}
			if s.FileType != nil {
				attrs = append(attrs, attribute.String("file.type", *s.FileType))
			}
			if s.FileSize != nil {
				attrs = append(attrs, attribute.Int64("file.size", *s.FileSize))
			}
			if s.Category != nil {
				attrs = append(attrs, attribute.String("file.category", *s.Category))
			}
			return attrs
		},
		func(s signal.UserFileSignal) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
			}
			if s.UID != nil {
				fields["file.uid"] = *s.UID
			}
			if s.UserUID != nil {
				fields["user.uid"] = *s.UserUID
			}
			if s.FileName != nil {
				fields["file.name"] = *s.FileName
			}
			if s.FileType != nil {
				fields["file.type"] = *s.FileType
			}
			if s.FileSize != nil {
				fields["file.size"] = *s.FileSize
			}
			if s.Category != nil {
				fields["file.category"] = *s.Category
			}
			return fields
		},
	)
}
