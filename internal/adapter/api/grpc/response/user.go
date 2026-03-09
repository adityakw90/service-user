package response

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// ToProtoUser converts domain User to proto User.
func ToProtoUser(u *model.User) *user.User {
	if u == nil {
		return nil
	}
	return &user.User{
		Uid:       u.UID,
		Username:  u.Username,
		Email:     u.Email,
		Status:    int32(u.Status),
		CreatedAt: toProtoTimestampPB(u.CreatedAt),
		UpdatedAt: toProtoTimestampPB(u.UpdatedAt),
		DeletedAt: toProtoTimestampPBPtr(u.DeletedAt),
	}
}

// toProtoTimestampPBPtr converts *time.Time to protobuf timestamp.
func toProtoTimestampPBPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return toProtoTimestampPB(*t)
}

// ToProtoUserList converts domain Users to proto ListResponse.
func ToProtoUserList(users *model.Users, meta *model.Meta) *user.ListResponse {
	if users == nil {
		return &user.ListResponse{Meta: ToProtoMeta(meta)}
	}

	items := make([]*user.User, len(users.Items))
	for i, u := range users.Items {
		items[i] = ToProtoUser(&u)
	}

	return &user.ListResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}
