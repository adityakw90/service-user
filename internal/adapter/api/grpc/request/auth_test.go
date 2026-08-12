package request

import (
	"testing"

	authpb "github.com/adityakw90/service-user-proto/gen/go/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAuthRequestFromPb(t *testing.T) {
	deviceIp := " 192.168.1.1 "
	deviceName := " Chrome "
	deviceFingerprint := " fingerprint-abc "
	emptyIp := "  "
	emptyName := " "
	emptyFingerprint := ""
	extraMap, err := structpb.NewStruct(map[string]any{"os": "Linux"})
	require.NoError(t, err)

	tests := []struct {
		name            string
		req             *authpb.AuthRequest
		wantIdentifier  string
		wantIdType      string
		wantPassword    string
		wantDeviceIP    *string
		wantDeviceName  *string
		wantFingerprint *string
		wantExtraKey    string
		wantExtraVal    any
	}{
		{
			name: "all optional fields populated and trimmed",
			req: &authpb.AuthRequest{
				Identifier:        "  john_doe  ",
				IdentifierType:    "  username  ",
				Password:          "  password123  ",
				DeviceIp:          &deviceIp,
				DeviceName:        &deviceName,
				DeviceFingerprint: &deviceFingerprint,
				Extra:             extraMap,
			},
			wantIdentifier:  "john_doe",
			wantIdType:      "username",
			wantPassword:    "password123",
			wantDeviceIP:    strPtr("192.168.1.1"),
			wantDeviceName:  strPtr("Chrome"),
			wantFingerprint: strPtr("fingerprint-abc"),
			wantExtraKey:    "os",
			wantExtraVal:    "Linux",
		},
		{
			name: "whitespace-only optional fields become nil",
			req: &authpb.AuthRequest{
				Identifier:        "john_doe",
				IdentifierType:    "username",
				Password:          "pass",
				DeviceIp:          &emptyIp,
				DeviceName:        &emptyName,
				DeviceFingerprint: &emptyFingerprint,
			},
			wantIdentifier:  "john_doe",
			wantIdType:      "username",
			wantPassword:    "pass",
			wantDeviceIP:    nil,
			wantDeviceName:  nil,
			wantFingerprint: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AuthRequestFromPb(tt.req)
			assert.Equal(t, tt.wantIdentifier, got.Identifier)
			assert.Equal(t, tt.wantIdType, got.IdentifierType)
			assert.Equal(t, tt.wantPassword, got.Password)
			assertOptionalString(t, tt.wantDeviceIP, got.DeviceIP, "DeviceIP")
			assertOptionalString(t, tt.wantDeviceName, got.DeviceName, "DeviceName")
			assertOptionalString(t, tt.wantFingerprint, got.DeviceFingerprint, "DeviceFingerprint")

			if tt.wantExtraKey != "" {
				require.NotNil(t, got.Extra)
				assert.Equal(t, tt.wantExtraVal, (*got.Extra)[tt.wantExtraKey])
			}

			// Also verify ToAuthParams maps all fields through unchanged.
			params := got.ToAuthParams()
			require.NotNil(t, params)
			assert.Equal(t, got.Identifier, params.Identifier)
			assert.Equal(t, got.IdentifierType, params.IdentifierType)
			assert.Equal(t, got.Password, params.Password)
			assert.Equal(t, got.DeviceIP, params.DeviceIP)
			assert.Equal(t, got.DeviceName, params.DeviceName)
			assert.Equal(t, got.DeviceFingerprint, params.DeviceFingerprint)
			assert.Equal(t, got.Extra, params.Extra)
		})
	}
}

func TestRefreshTokenRequestFromPb(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantToken string
	}{
		{name: "trims surrounding whitespace", raw: "  refresh-token-123  ", wantToken: "refresh-token-123"},
		{name: "already trimmed value unchanged", raw: "refresh-token-abc", wantToken: "refresh-token-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RefreshTokenRequestFromPb(&authpb.RefreshTokenRequest{RefreshToken: tt.raw})
			assert.Equal(t, tt.wantToken, got.RefreshToken)
		})
	}
}

func TestValidateTokenRequestFromPb(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantToken string
	}{
		{name: "trims surrounding whitespace", raw: "  access-token-123  ", wantToken: "access-token-123"},
		{name: "already trimmed value unchanged", raw: "access-token-abc", wantToken: "access-token-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateTokenRequestFromPb(&authpb.ValidateTokenRequest{AccessToken: tt.raw})
			assert.Equal(t, tt.wantToken, got.AccessToken)
		})
	}
}

func TestVerifyPinRequestFromPb(t *testing.T) {
	tests := []struct {
		name     string
		uid      string
		code     string
		wantUid  string
		wantCode string
	}{
		{
			name:     "trims uid and code",
			uid:      "  user-uid  ",
			code:     "  123456  ",
			wantUid:  "user-uid",
			wantCode: "123456",
		},
		{
			name:     "already trimmed values unchanged",
			uid:      "user-uid-2",
			code:     "654321",
			wantUid:  "user-uid-2",
			wantCode: "654321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyPinRequestFromPb(&authpb.VerifyPinRequest{Uid: tt.uid, Code: tt.code})
			assert.Equal(t, tt.wantUid, got.Uid)
			assert.Equal(t, tt.wantCode, got.Code)
		})
	}
}

func TestGoogleOAuthRequestFromPb(t *testing.T) {
	tests := []struct {
		name            string
		raw             string
		wantRedirectUri string
	}{
		{name: "trims redirect URI", raw: "  http://localhost:3000/callback  ", wantRedirectUri: "http://localhost:3000/callback"},
		{name: "already trimmed value unchanged", raw: "https://example.com/cb", wantRedirectUri: "https://example.com/cb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GoogleOAuthRequestFromPb(&authpb.GoogleOAuthRequest{RedirectUri: tt.raw})
			assert.Equal(t, tt.wantRedirectUri, got.RedirectUri)
		})
	}
}

func TestHandleGoogleOAuthRequestFromPb(t *testing.T) {
	tests := []struct {
		name            string
		code            string
		state           string
		redirectUri     string
		wantCode        string
		wantState       string
		wantRedirectUri string
	}{
		{
			name:            "trims all fields",
			code:            "  oauth-code-abc  ",
			state:           "  oauth-state-123  ",
			redirectUri:     "  http://localhost:3000/callback  ",
			wantCode:        "oauth-code-abc",
			wantState:       "oauth-state-123",
			wantRedirectUri: "http://localhost:3000/callback",
		},
		{
			name:            "already trimmed values unchanged",
			code:            "code-xyz",
			state:           "state-xyz",
			redirectUri:     "https://example.com/cb",
			wantCode:        "code-xyz",
			wantState:       "state-xyz",
			wantRedirectUri: "https://example.com/cb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleGoogleOAuthRequestFromPb(&authpb.HandleGoogleOAuthRequest{
				Code:        tt.code,
				State:       tt.state,
				RedirectUri: tt.redirectUri,
			})
			assert.Equal(t, tt.wantCode, got.Code)
			assert.Equal(t, tt.wantState, got.State)
			assert.Equal(t, tt.wantRedirectUri, got.RedirectUri)
		})
	}
}

func TestRevokeTokenRequestFromPb(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		tokenType     string
		wantToken     string
		wantTokenType string
	}{
		{
			name:          "trims token and type",
			token:         "  token-to-revoke  ",
			tokenType:     "  refresh  ",
			wantToken:     "token-to-revoke",
			wantTokenType: "refresh",
		},
		{
			name:          "already trimmed values unchanged",
			token:         "token-abc",
			tokenType:     "access",
			wantToken:     "token-abc",
			wantTokenType: "access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RevokeTokenRequestFromPb(&authpb.RevokeTokenRequest{
				Token:     tt.token,
				TokenType: tt.tokenType,
			})
			assert.Equal(t, tt.wantToken, got.Token)
			assert.Equal(t, tt.wantTokenType, got.TokenType)
		})
	}
}

// strPtr is a test helper that returns a pointer to a string literal.
func strPtr(s string) *string { return &s }

// assertOptionalString checks a *string field: if want is nil the field must be nil,
// otherwise it must equal the pointed-to value.
func assertOptionalString(t *testing.T, want, got *string, field string) {
	t.Helper()
	if want == nil {
		assert.Nil(t, got, field+" should be nil")
	} else {
		require.NotNil(t, got, field+" should not be nil")
		assert.Equal(t, *want, *got, field)
	}
}
