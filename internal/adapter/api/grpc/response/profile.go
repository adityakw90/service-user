package response

import (
	"github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// ToProtoProfile converts domain UserProfile to proto Profile.
func ToProtoProfile(p *model.UserProfile) *user.Profile {
	if p == nil {
		return nil
	}
	return &user.Profile{
		Uid:        p.UserUID,
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		Bio:        p.Bio,
		Attributes: ToStruct(p.Attributes),
	}
}
