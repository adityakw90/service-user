package response

import (
	"testing"

	auth "github.com/adityakw90/service-user-proto/gen/go/auth"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

func TestToProtoToken(t *testing.T) {
	tests := []struct {
		name  string
		input *model.Token
		want  *auth.Token
	}{
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "Valid token",
			input: &model.Token{
				Access:  "access-token-123",
				Refresh: "refresh-token-456",
			},
			want: &auth.Token{
				AccessToken:  "access-token-123",
				RefreshToken: "refresh-token-456",
			},
		},
		{
			name: "Token with empty values",
			input: &model.Token{
				Access:  "",
				Refresh: "",
			},
			want: &auth.Token{
				AccessToken:  "",
				RefreshToken: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoToken(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("ToProtoToken() = %v, want nil", got)
				}
				return
			}

			if got.AccessToken != tt.want.AccessToken {
				t.Errorf("ToProtoToken().AccessToken = %v, want %v", got.AccessToken, tt.want.AccessToken)
			}
			if got.RefreshToken != tt.want.RefreshToken {
				t.Errorf("ToProtoToken().RefreshToken = %v, want %v", got.RefreshToken, tt.want.RefreshToken)
			}
		})
	}
}
