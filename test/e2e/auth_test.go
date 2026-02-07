package e2e

import (
	"context"
	"testing"

	authgrpc "github.com/adityakw90/service-user-proto/gen/go/auth"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/pkg/util"
	testutil "github.com/adityakw90/service-user/test/util"
	"github.com/stretchr/testify/require"
)

/*
	testing all auth grpc service
	each test function is reflecting the service endpoint
*/

// TestE2E_AuthService_Auth tests authentication endpoint
func TestE2E_AuthService_Auth(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		authReq    func(t *testing.T) *authgrpc.AuthRequest
		wantErr    bool
		errMsg     string
		verifyFunc func(t *testing.T, token *authgrpc.Token)
	}{
		{
			name: "Auth with email and correct password",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "authemail", "authemail@example.com", "Password123!")
			},
			authReq: func(t *testing.T) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        "authemail@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("test-device-fingerprint"),
					DeviceName:        util.Ptr("Test Device"),
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, token *authgrpc.Token) {
				require.NotNil(t, token)
				require.NotEmpty(t, token.AccessToken)
				require.NotEmpty(t, token.RefreshToken)
			},
		},
		{
			name: "Auth with username and correct password",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "authuser", "authuser@example.com", "Password123!")
			},
			authReq: func(t *testing.T) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        "authuser",
					IdentifierType:    "username",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("test-device-fingerprint-2"),
					DeviceName:        util.Ptr("Test Device 2"),
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, token *authgrpc.Token) {
				require.NotNil(t, token)
				require.NotEmpty(t, token.AccessToken)
			},
		},
		{
			name: "Auth with invalid credentials",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "invalidcredse2e", "invalidcredse2e@example.com", "ValidPassword123!")
			},
			authReq: func(t *testing.T) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        "invalidcredse2e@example.com",
					IdentifierType:    "email",
					Password:          "WrongPassword123!",
					DeviceFingerprint: util.Ptr("test-device-fingerprint-3"),
				}
			},
			wantErr: true,
		},
		{
			name: "Auth with empty identifier type",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "emptytype", "emptytype@example.com", "Password123!")
			},
			authReq: func(t *testing.T) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        "emptytype@example.com",
					IdentifierType:    "",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("test-device"),
					DeviceName:        util.Ptr("test"),
				}
			},
			wantErr: true,
		},
		{
			name: "Auth with non-existent user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "nonexistent"
			},
			authReq: func(t *testing.T) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        "nonexistent@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("test-device"),
					DeviceName:        util.Ptr("test"),
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, grpcClient, cleanup := setupE2ETest(t)
			defer cleanup()

			ctx := context.Background()
			_ = tt.setup(t, grpcClient)
			authReq := tt.authReq(t)

			token, err := grpcClient.AuthClient.Auth(ctx, authReq)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, token)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, token)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, token)
				}
			}
		})
	}
}

// TestE2E_AuthService_RefreshToken tests refresh token endpoint
func TestE2E_AuthService_RefreshToken(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) *authgrpc.Token
		refreshReq func(t *testing.T, grpcClient *testutil.TestGRPCClient, initialToken *authgrpc.Token) *authgrpc.RefreshTokenRequest
		wantErr    bool
		errMsg     string
		verifyFunc func(t *testing.T, initialToken, newToken *authgrpc.Token)
	}{
		{
			name: "Refresh token with valid token",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) *authgrpc.Token {
				createTestUser(t, grpcClient, "refreshtokene2e", "refreshtokene2e@example.com", "RefreshToken123!")
				ctx := context.Background()
				authReq := &authgrpc.AuthRequest{
					Identifier:        "refreshtokene2e@example.com",
					IdentifierType:    "email",
					Password:          "RefreshToken123!",
					DeviceFingerprint: util.Ptr("refresh-device-fingerprint"),
					DeviceName:        util.Ptr("Refresh Device"),
				}
				token, err := grpcClient.AuthClient.Auth(ctx, authReq)
				require.NoError(t, err)
				return token
			},
			refreshReq: func(t *testing.T, grpcClient *testutil.TestGRPCClient, initialToken *authgrpc.Token) *authgrpc.RefreshTokenRequest {
				return &authgrpc.RefreshTokenRequest{
					RefreshToken: initialToken.RefreshToken,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, initialToken, newToken *authgrpc.Token) {
				require.NotEmpty(t, newToken.AccessToken)
				require.NotEmpty(t, newToken.RefreshToken)
				require.NotEqual(t, initialToken.AccessToken, newToken.AccessToken)
			},
		},
		{
			name: "Refresh token twice (token rotation)",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) *authgrpc.Token {
				createTestUser(t, grpcClient, "refresh2x", "refresh2x@example.com", "Password123!")
				ctx := context.Background()
				authReq := &authgrpc.AuthRequest{
					Identifier:        "refresh2x@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("test-device-2x"),
					DeviceName:        util.Ptr("Test Device"),
				}
				token, err := grpcClient.AuthClient.Auth(ctx, authReq)
				require.NoError(t, err)
				return token
			},
			refreshReq: func(t *testing.T, grpcClient *testutil.TestGRPCClient, initialToken *authgrpc.Token) *authgrpc.RefreshTokenRequest {
				// First refresh
				ctx := context.Background()
				firstRefresh, err := grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
					RefreshToken: initialToken.RefreshToken,
				})
				require.NoError(t, err)

				// Return second refresh request
				return &authgrpc.RefreshTokenRequest{
					RefreshToken: firstRefresh.RefreshToken,
				}
			},
			wantErr: false,
		},
		{
			name: "Refresh token with invalid/expired token",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) *authgrpc.Token {
				createTestUser(t, grpcClient, "expiredtoken", "expiredtoken@example.com", "Password123!")
				ctx := context.Background()
				authReq := &authgrpc.AuthRequest{
					Identifier:        "expiredtoken@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("test-device-exp"),
					DeviceName:        util.Ptr("Test Device"),
				}
				token, err := grpcClient.AuthClient.Auth(ctx, authReq)
				require.NoError(t, err)
				return token
			},
			refreshReq: func(t *testing.T, grpcClient *testutil.TestGRPCClient, initialToken *authgrpc.Token) *authgrpc.RefreshTokenRequest {
				return &authgrpc.RefreshTokenRequest{
					RefreshToken: "invalid.refresh.token",
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, grpcClient, cleanup := setupE2ETest(t)
			defer cleanup()

			ctx := context.Background()
			initialToken := tt.setup(t, grpcClient)
			refreshReq := tt.refreshReq(t, grpcClient, initialToken)

			newToken, err := grpcClient.AuthClient.RefreshToken(ctx, refreshReq)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, newToken)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, initialToken, newToken)
				}
			}
		})
	}
}

// TestE2E_AuthService_ValidateToken tests validate token endpoint
func TestE2E_AuthService_ValidateToken(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string)
		tokenReq   func(t *testing.T, accessToken string) *authgrpc.ValidateTokenRequest
		wantErr    bool
		errMsg     string
		verifyFunc func(t *testing.T, claims *authgrpc.ValidateTokenResponse, uid, email string)
	}{
		{
			name: "Validate valid token",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string) {
				email := "validatetokene2e@example.com"
				uid := createTestUser(t, grpcClient, "validatetokene2e", email, "ValidateToken123!")
				return uid, email
			},
			tokenReq: func(t *testing.T, accessToken string) *authgrpc.ValidateTokenRequest {
				return &authgrpc.ValidateTokenRequest{
					AccessToken: accessToken,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, claims *authgrpc.ValidateTokenResponse, uid, email string) {
				require.Equal(t, uid, claims.Uid)
				require.Equal(t, email, claims.Identifier)
			},
		},
		{
			name: "Validate invalid token",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string) {
				return "", ""
			},
			tokenReq: func(t *testing.T, accessToken string) *authgrpc.ValidateTokenRequest {
				return &authgrpc.ValidateTokenRequest{
					AccessToken: "invalid.token.here",
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, grpcClient, cleanup := setupE2ETest(t)
			defer cleanup()

			ctx := context.Background()
			uid, email := tt.setup(t, grpcClient)

			var accessToken string
			if uid != "" {
				authReq := &authgrpc.AuthRequest{
					Identifier:        email,
					IdentifierType:    "email",
					Password:          "ValidateToken123!",
					DeviceFingerprint: util.Ptr("validate-device-fingerprint"),
					DeviceName:        util.Ptr("Validate Device"),
				}
				token, err := grpcClient.AuthClient.Auth(ctx, authReq)
				require.NoError(t, err)
				accessToken = token.AccessToken
			}

			validateReq := tt.tokenReq(t, accessToken)

			claims, err := grpcClient.AuthClient.ValidateToken(ctx, validateReq)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, claims)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, claims, uid, email)
				}
			}
		})
	}
}

// TestE2E_AuthService_RevokeToken tests revoke token endpoint
func TestE2E_AuthService_RevokeToken(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) *authgrpc.Token
		revokeReq  func(t *testing.T, token *authgrpc.Token) *authgrpc.RevokeTokenRequest
		wantErr    bool
		verifyFunc func(t *testing.T, grpcClient *testutil.TestGRPCClient, token *authgrpc.Token)
	}{
		{
			name: "Revoke refresh token",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) *authgrpc.Token {
				createTestUser(t, grpcClient, "revoketoken", "revoketoken@example.com", "Password123!")
				ctx := context.Background()
				authReq := &authgrpc.AuthRequest{
					Identifier:        "revoketoken@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("test-device-revoke"),
					DeviceName:        util.Ptr("Test Device"),
				}
				token, err := grpcClient.AuthClient.Auth(ctx, authReq)
				require.NoError(t, err)
				return token
			},
			revokeReq: func(t *testing.T, token *authgrpc.Token) *authgrpc.RevokeTokenRequest {
				return &authgrpc.RevokeTokenRequest{
					Token:     token.RefreshToken,
					TokenType: "refresh",
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, grpcClient *testutil.TestGRPCClient, token *authgrpc.Token) {
				ctx := context.Background()
				_, err := grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
					RefreshToken: token.RefreshToken,
				})
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, grpcClient, cleanup := setupE2ETest(t)
			defer cleanup()

			ctx := context.Background()
			token := tt.setup(t, grpcClient)
			revokeReq := tt.revokeReq(t, token)

			_, err := grpcClient.AuthClient.RevokeToken(ctx, revokeReq)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, grpcClient, token)
				}
			}
		})
	}
}

// TestE2E_AuthService_VerifyPin tests verify PIN endpoint
func TestE2E_AuthService_VerifyPin(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		verifyReq func(t *testing.T) *authgrpc.VerifyPinRequest
		wantErr   bool
		wantValid bool
	}{
		{
			name: "Verify correct PIN",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				uid := createTestUser(t, grpcClient, "pinusere2e", "pinusere2e@example.com", "PinPassword123!")
				ctx := context.Background()
				_, err := grpcClient.UserClient.UpdatePin(ctx, &usergrpc.UpdatePinRequest{
					UserUid: uid,
					Pin:     "529183",
				})
				require.NoError(t, err)
				return uid
			},
			verifyReq: func(t *testing.T) *authgrpc.VerifyPinRequest {
				return &authgrpc.VerifyPinRequest{
					Uid:  "", // Will be set in test
					Code: "529183",
				}
			},
			wantErr:   false,
			wantValid: true,
		},
		{
			name: "Verify incorrect PIN",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				uid := createTestUser(t, grpcClient, "invalidpine2e", "invalidpine2e@example.com", "InvalidPin123!")
				ctx := context.Background()
				_, err := grpcClient.UserClient.UpdatePin(ctx, &usergrpc.UpdatePinRequest{
					UserUid: uid,
					Pin:     "529183",
				})
				require.NoError(t, err)
				return uid
			},
			verifyReq: func(t *testing.T) *authgrpc.VerifyPinRequest {
				return &authgrpc.VerifyPinRequest{
					Uid:  "", // Will be set in test
					Code: "718294",
				}
			},
			wantErr:   false,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, grpcClient, cleanup := setupE2ETest(t)
			defer cleanup()

			ctx := context.Background()
			uid := tt.setup(t, grpcClient)
			verifyReq := tt.verifyReq(t)
			verifyReq.Uid = uid

			verifyResp, err := grpcClient.AuthClient.VerifyPin(ctx, verifyReq)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantValid, verifyResp.Valid)
			}
		})
	}
}
