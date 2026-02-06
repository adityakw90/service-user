package request

import (
	"strings"

	auth "github.com/adityakw90/service-user-proto/gen/go/auth"
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
