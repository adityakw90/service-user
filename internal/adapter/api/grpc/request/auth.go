package request

import (
	"strings"

	auth "github.com/adityakw90/service-user-proto/gen/go/auth"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

// AuthRequest represents validated auth request data.
type AuthRequest struct {
	Identifier        string          `validate:"required"`
	IdentifierType    string          `validate:"required,oneof=email username phone"`
	Password          string          `validate:"required"`
	DeviceFingerprint *string         `validate:"omitempty,max=255"`
	DeviceName        *string         `validate:"omitempty,max=100"`
	DeviceIP          *string         `validate:"omitempty,max=100"`
	Extra             *map[string]any `validate:"omitempty"`
}

func (r *AuthRequest) ToAuthParams() *params.AuthParams {
	return &params.AuthParams{
		Identifier:        r.Identifier,
		IdentifierType:    r.IdentifierType,
		Password:          r.Password,
		DeviceFingerprint: r.DeviceFingerprint,
		DeviceName:        r.DeviceName,
		DeviceIP:          r.DeviceIP,
		Extra:             r.Extra,
	}
}

func AuthRequestFromPb(req *auth.AuthRequest) *AuthRequest {
	payload := &AuthRequest{
		Identifier:     strings.TrimSpace(req.GetIdentifier()),
		IdentifierType: strings.TrimSpace(req.GetIdentifierType()),
		Password:       strings.TrimSpace(req.GetPassword()),
	}

	if req.DeviceIp != nil {
		fieldDeviceIP := strings.TrimSpace(req.GetDeviceIp())
		if fieldDeviceIP != "" {
			payload.DeviceIP = &fieldDeviceIP
		}
	}

	if req.DeviceName != nil {
		fieldDeviceName := strings.TrimSpace(req.GetDeviceName())
		if fieldDeviceName != "" {
			payload.DeviceName = &fieldDeviceName
		}
	}

	if req.DeviceFingerprint != nil {
		fieldDeviceFingerprint := strings.TrimSpace(req.GetDeviceFingerprint())
		if fieldDeviceFingerprint != "" {
			payload.DeviceFingerprint = &fieldDeviceFingerprint
		}
	}

	if req.Extra != nil {
		fieldExtra := req.GetExtra().AsMap()
		payload.Extra = &fieldExtra
	}

	return payload
}

// RefreshTokenRequest represents validated refresh token request.
type RefreshTokenRequest struct {
	RefreshToken string `validate:"required"`
}

// RefreshTokenRequestFromPb creates a RefreshTokenRequest from protobuf.
func RefreshTokenRequestFromPb(req *auth.RefreshTokenRequest) *RefreshTokenRequest {
	return &RefreshTokenRequest{
		RefreshToken: strings.TrimSpace(req.RefreshToken),
	}
}

// ValidateTokenRequest represents validated token validation request.
type ValidateTokenRequest struct {
	AccessToken string `validate:"required"`
}

// ValidateTokenRequestFromPb creates a ValidateTokenRequest from protobuf.
func ValidateTokenRequestFromPb(req *auth.ValidateTokenRequest) *ValidateTokenRequest {
	return &ValidateTokenRequest{
		AccessToken: strings.TrimSpace(req.AccessToken),
	}
}

// VerifyPinRequest represents validated PIN verification request.
type VerifyPinRequest struct {
	Uid  string `validate:"required"`
	Code string `validate:"required,pin"`
}

// VerifyPinRequestFromPb creates a VerifyPinRequest from protobuf.
func VerifyPinRequestFromPb(req *auth.VerifyPinRequest) *VerifyPinRequest {
	return &VerifyPinRequest{
		Uid:  strings.TrimSpace(req.Uid),
		Code: strings.TrimSpace(req.Code),
	}
}

// GoogleOAuthRequest represents validated Google OAuth request.
type GoogleOAuthRequest struct {
	RedirectUri string `validate:"required,uri"`
}

// GoogleOAuthRequestFromPb creates a GoogleOAuthRequest from protobuf.
func GoogleOAuthRequestFromPb(req *auth.GoogleOAuthRequest) *GoogleOAuthRequest {
	return &GoogleOAuthRequest{
		RedirectUri: strings.TrimSpace(req.RedirectUri),
	}
}

// HandleGoogleOAuthRequest represents validated OAuth callback request.
type HandleGoogleOAuthRequest struct {
	Code        string `validate:"required"`
	State       string `validate:"required"`
	RedirectUri string `validate:"required,uri"`
}

// HandleGoogleOAuthRequestFromPb creates a HandleGoogleOAuthRequest from protobuf.
func HandleGoogleOAuthRequestFromPb(req *auth.HandleGoogleOAuthRequest) *HandleGoogleOAuthRequest {
	return &HandleGoogleOAuthRequest{
		Code:        strings.TrimSpace(req.GetCode()),
		State:       strings.TrimSpace(req.GetState()),
		RedirectUri: strings.TrimSpace(req.GetRedirectUri()),
	}
}

// RevokeTokenRequest represents validated token revocation request.
type RevokeTokenRequest struct {
	Token     string `validate:"required"`
	TokenType string `validate:"required,oneof=access refresh"`
}

// RevokeTokenRequestFromPb creates a RevokeTokenRequest from protobuf.
func RevokeTokenRequestFromPb(req *auth.RevokeTokenRequest) *RevokeTokenRequest {
	return &RevokeTokenRequest{
		Token:     strings.TrimSpace(req.Token),
		TokenType: strings.TrimSpace(req.TokenType),
	}
}
