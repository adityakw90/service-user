package response

import (
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	userproto "github.com/adityakw90/service-user-proto/gen/go/user"
)

func TestToProtoUser(t *testing.T) {
	now := time.Now()
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name  string
		input *model.User
		want  *userproto.User
	}{
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "Valid user",
			input: &model.User{
				UID:       uid.String(),
				Username:  "testuser",
				Email:     "test@example.com",
				Status:    model.UserStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
			want: &userproto.User{
				Uid:       "00000000-0000-0000-0000-000000000001",
				Username:  "testuser",
				Email:     "test@example.com",
				Status:    int32(model.UserStatusActive),
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
		},
		{
			name: "User with deleted timestamp",
			input: &model.User{
				UID:       uid.String(),
				Username:  "deleteduser",
				Email:     "deleted@example.com",
				Status:    model.UserStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
				DeletedAt: &now,
			},
			want: &userproto.User{
				Uid:       "00000000-0000-0000-0000-000000000001",
				Username:  "deleteduser",
				Email:     "deleted@example.com",
				Status:    int32(model.UserStatusActive),
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
				DeletedAt: timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoUser(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("ToProtoUser() = %v, want nil", got)
				}
				return
			}

			if got.Uid != tt.want.Uid {
				t.Errorf("ToProtoUser().Uid = %v, want %v", got.Uid, tt.want.Uid)
			}
			if got.Username != tt.want.Username {
				t.Errorf("ToProtoUser().Username = %v, want %v", got.Username, tt.want.Username)
			}
			if got.Email != tt.want.Email {
				t.Errorf("ToProtoUser().Email = %v, want %v", got.Email, tt.want.Email)
			}
			if got.Status != tt.want.Status {
				t.Errorf("ToProtoUser().Status = %v, want %v", got.Status, tt.want.Status)
			}
		})
	}
}

func TestToProtoUserList(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name     string
		users    *model.Users
		meta     *model.Meta
		wantLen  int
		wantPage int32
	}{
		{
			name: "Empty list",
			users: &model.Users{
				Items: []model.User{},
			},
			meta:     &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
			wantLen:  0,
			wantPage: 1,
		},
		{
			name: "List with users",
			users: &model.Users{
				Items: []model.User{
					{UID: uid.String(), Username: "user1", Email: "user1@example.com", Status: model.UserStatusActive},
					{UID: uid.String(), Username: "user2", Email: "user2@example.com", Status: model.UserStatusActive},
				},
			},
			meta:     &model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
			wantLen:  2,
			wantPage: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoUserList(tt.users, tt.meta)

			if len(got.Items) != tt.wantLen {
				t.Errorf("ToProtoUserList() len = %d, want %d", len(got.Items), tt.wantLen)
			}
			if got.Meta.Page != tt.wantPage {
				t.Errorf("ToProtoUserList().Meta.Page = %d, want %d", got.Meta.Page, tt.wantPage)
			}
		})
	}
}
