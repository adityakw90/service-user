package response

import (
	"testing"
	"time"

	"github.com/adityakw90/service-user-proto/gen/go/common"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

func TestToProtoMeta(t *testing.T) {
	tests := []struct {
		name  string
		input *model.Meta
		want  *common.Meta
	}{
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "Valid meta",
			input: &model.Meta{
				Page:  1,
				Limit: 10,
				Total: 100,
				Pages: 10,
			},
			want: &common.Meta{
				Page:  1,
				Limit: 10,
				Total: 100,
				Pages: 10,
			},
		},
		{
			name: "Meta with zero values",
			input: &model.Meta{
				Page:  0,
				Limit: 0,
				Total: 0,
				Pages: 0,
			},
			want: &common.Meta{
				Page:  0,
				Limit: 0,
				Total: 0,
				Pages: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoMeta(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("ToProtoMeta() = %v, want nil", got)
				}
				return
			}

			if got.Page != tt.want.Page {
				t.Errorf("ToProtoMeta().Page = %v, want %v", got.Page, tt.want.Page)
			}
			if got.Limit != tt.want.Limit {
				t.Errorf("ToProtoMeta().Limit = %v, want %v", got.Limit, tt.want.Limit)
			}
			if got.Total != tt.want.Total {
				t.Errorf("ToProtoMeta().Total = %v, want %v", got.Total, tt.want.Total)
			}
			if got.Pages != tt.want.Pages {
				t.Errorf("ToProtoMeta().Pages = %v, want %v", got.Pages, tt.want.Pages)
			}
		})
	}
}

func TestToProtoTimestampPB(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
		want  interface{}
	}{
		{
			name:  "Zero time",
			input: time.Time{},
			want:  nil,
		},
		{
			name:  "Valid time",
			input: time.Date(2024, 3, 9, 12, 0, 0, 0, time.UTC),
			want:  "non-nil timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toProtoTimestampPB(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("toProtoTimestampPB() = %v, want nil", got)
				}
			} else if got == nil {
				t.Errorf("toProtoTimestampPB() = nil, want non-nil")
			}
		})
	}
}

func TestToStruct(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		want  interface{}
	}{
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "Valid map",
			input: map[string]any{
				"key1": "value1",
				"key2": 123,
			},
			want: "non-nil struct",
		},
		{
			name:  "Empty map",
			input: map[string]any{},
			want:  "non-nil struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToStruct(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("ToStruct() = %v, want nil", got)
				}
			} else if got == nil {
				t.Errorf("ToStruct() = nil, want non-nil")
			}
		})
	}
}
