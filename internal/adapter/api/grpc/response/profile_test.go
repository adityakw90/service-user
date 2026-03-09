package response

import (
	"testing"

	"github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

func TestToProtoProfile(t *testing.T) {
	tests := []struct {
		name  string
		input *model.UserProfile
		want  *user.Profile
	}{
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "Valid profile",
			input: &model.UserProfile{
				UserUID:    "user-123",
				FirstName:  "John",
				LastName:   "Doe",
				Bio:        "Test bio",
				Attributes: map[string]any{"key1": "value1", "key2": 123},
			},
			want: &user.Profile{
				Uid:       "user-123",
				FirstName: "John",
				LastName:  "Doe",
				Bio:       "Test bio",
			},
		},
		{
			name: "Profile with nil attributes",
			input: &model.UserProfile{
				UserUID:    "user-456",
				FirstName:  "Jane",
				LastName:   "Smith",
				Bio:        "",
				Attributes: nil,
			},
			want: &user.Profile{
				Uid:        "user-456",
				FirstName:  "Jane",
				LastName:   "Smith",
				Bio:        "",
				Attributes: nil,
			},
		},
		{
			name: "Profile with empty attributes map",
			input: &model.UserProfile{
				UserUID:    "user-789",
				FirstName:  "Bob",
				LastName:   "Johnson",
				Bio:        "Developer",
				Attributes: map[string]any{},
			},
			want: &user.Profile{
				Uid:       "user-789",
				FirstName: "Bob",
				LastName:  "Johnson",
				Bio:       "Developer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoProfile(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("ToProtoProfile() = %v, want nil", got)
				}
				return
			}

			if got.FirstName != tt.want.FirstName {
				t.Errorf("ToProtoProfile().FirstName = %v, want %v", got.FirstName, tt.want.FirstName)
			}
			if got.LastName != tt.want.LastName {
				t.Errorf("ToProtoProfile().LastName = %v, want %v", got.LastName, tt.want.LastName)
			}
			if got.Bio != tt.want.Bio {
				t.Errorf("ToProtoProfile().Bio = %v, want %v", got.Bio, tt.want.Bio)
			}
			if got.Uid != tt.want.Uid {
				t.Errorf("ToProtoProfile().Uid = %v, want %v", got.Uid, tt.want.Uid)
			}
			// Verify Attributes field is correctly converted
			if tt.input.Attributes != nil && got.Attributes == nil {
				t.Errorf("ToProtoProfile().Attributes = nil, want non-nil (conversion failed)")
			}
			if tt.input.Attributes == nil && got.Attributes != nil {
				t.Errorf("ToProtoProfile().Attributes = %v, want nil", got.Attributes)
			}
		})
	}
}
