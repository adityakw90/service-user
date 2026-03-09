package response

import (
	"time"

	"github.com/adityakw90/service-user-proto/gen/go/common"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// ToProtoMeta converts domain meta to proto meta.
// Note: Page, Limit, and Pages are expected to be within int32 range for pagination.
// Values exceeding int32 max will overflow, which should be validated at the service layer.
func ToProtoMeta(m *model.Meta) *common.Meta {
	if m == nil {
		return nil
	}
	return &common.Meta{
		Page:  int32(m.Page),
		Limit: int32(m.Limit),
		Total: m.Total,
		Pages: int32(m.Pages),
	}
}

// toProtoTimestampPB converts time.Time to protobuf timestamp.
func toProtoTimestampPB(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// ToStruct converts map[string]any to protobuf Struct.
// Returns nil if the input map is nil or if conversion fails (e.g., contains invalid data types).
func ToStruct(m map[string]any) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		// Return nil if conversion fails - map contains invalid data
		return nil
	}
	return s
}
