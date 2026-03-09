package response

import (
	auth "github.com/adityakw90/service-user-proto/gen/go/auth"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// ToProtoToken converts domain Token to proto auth.Token.
func ToProtoToken(t *model.Token) *auth.Token {
	if t == nil {
		return nil
	}
	return &auth.Token{
		AccessToken:  t.Access,
		RefreshToken: t.Refresh,
	}
}
