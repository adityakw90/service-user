package response

import (
	"time"

	"github.com/adityakw90/service-user-proto/gen/go/common"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// ToProtoMeta converts domain meta to proto meta.
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

// toProtoTimestamp converts time.Time to protobuf timestamp.
func toProtoTimestamp(t time.Time) *time.Time {
	return &t
}

// toProtoTimestampPB converts time.Time to protobuf timestamp.
func toProtoTimestampPB(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return &timestamppb.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// ToStruct converts map[string]any to protobuf Struct.
func ToStruct(m map[string]any) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, _ := structpb.NewStruct(m)
	return s
}
