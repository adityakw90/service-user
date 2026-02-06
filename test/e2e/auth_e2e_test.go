package e2e

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	authgrpc "github.com/adityakw90/service-user-proto/gen/go/auth"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/stretchr/testify/require"
)

func TestUserRegistration(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	req := &usergrpc.AddRequest{
		Username: "testuser_e2e",
		Email:    "testuser_e2e@example.com",
		Password: "SecurePassword123!",
	}

	resp, err := grpcClient.UserClient.Add(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Uid)

	// Verify user was created by getting it
	userResp, err := grpcClient.UserClient.Get(ctx, &usergrpc.GetRequest{Uid: resp.Uid})
	require.NoError(t, err)
	require.Equal(t, req.Username, userResp.Username)
	require.Equal(t, req.Email, userResp.Email)
	require.Equal(t, int32(model.UserStatusActive), userResp.Status)
}

func TestUserRegistrationDuplicateEmail(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	email := "duplicate_e2e@example.com"

	// Create first user
	req1 := &usergrpc.AddRequest{
		Username: "user1_e2e",
		Email:    email,
		Password: "Password123!",
	}
	_, err := grpcClient.UserClient.Add(ctx, req1)
	require.NoError(t, err)

	// Try to create second user with same email
	req2 := &usergrpc.AddRequest{
		Username: "user2_e2e",
		Email:    email,
		Password: "DifferentPassword123!",
	}

	_, err = grpcClient.UserClient.Add(ctx, req2)
	require.Error(t, err)
}

func TestUserLogin(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	password := "LoginPassword123!"
	createReq := &usergrpc.AddRequest{
		Username: "loginuser_e2e",
		Email:    "loginuser_e2e@example.com",
		Password: password,
	}

	_, err := grpcClient.UserClient.Add(ctx, createReq)
	require.NoError(t, err)

	// Login with email
	authReq := &authgrpc.AuthRequest{
		Identifier:        createReq.Email,
		IdentifierType:    "email",
		Password:          password,
		DeviceFingerprint: "test-device-fingerprint",
		DeviceName:        "Test Device",
	}

	token, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, token.AccessToken)
	require.NotEmpty(t, token.RefreshToken)
}

func TestUserLoginWithUsername(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	password := "UsernameLogin123!"
	createReq := &usergrpc.AddRequest{
		Username: "username_login_e2e",
		Email:    "username_login_e2e@example.com",
		Password: password,
	}

	_, err := grpcClient.UserClient.Add(ctx, createReq)
	require.NoError(t, err)

	// Login with username
	authReq := &authgrpc.AuthRequest{
		Identifier:        createReq.Username,
		IdentifierType:    "username",
		Password:          password,
		DeviceFingerprint: "test-device-fingerprint-2",
		DeviceName:        "Test Device 2",
	}

	token, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, token.AccessToken)
}

func TestUserLoginInvalidCredentials(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	password := "ValidPassword123!"
	createReq := &usergrpc.AddRequest{
		Username: "invalidcreds_e2e",
		Email:    "invalidcreds_e2e@example.com",
		Password: password,
	}

	_, err := grpcClient.UserClient.Add(ctx, createReq)
	require.NoError(t, err)

	// Try to login with wrong password
	authReq := &authgrpc.AuthRequest{
		Identifier:        createReq.Email,
		IdentifierType:    "email",
		Password:          "WrongPassword123!",
		DeviceFingerprint: "test-device-fingerprint-3",
	}

	_, err = grpcClient.AuthClient.Auth(ctx, authReq)
	require.Error(t, err)
}

func TestRefreshToken(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	password := "RefreshToken123!"
	createReq := &usergrpc.AddRequest{
		Username: "refreshtoken_e2e",
		Email:    "refreshtoken_e2e@example.com",
		Password: password,
	}

	_, err := grpcClient.UserClient.Add(ctx, createReq)
	require.NoError(t, err)

	authReq := &authgrpc.AuthRequest{
		Identifier:        createReq.Email,
		IdentifierType:    "email",
		Password:          password,
		DeviceFingerprint: "refresh-device-fingerprint",
		DeviceName:        "Refresh Device",
	}

	// Get initial tokens
	initialToken, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)
	require.NotNil(t, initialToken)

	// When - refresh token
	newToken, err := grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
		RefreshToken: initialToken.RefreshToken,
	})

	require.NoError(t, err)
	require.NotNil(t, newToken)
	require.NotEmpty(t, newToken.AccessToken)
	require.NotEmpty(t, newToken.RefreshToken)
	// New access token should be different
	require.NotEqual(t, initialToken.AccessToken, newToken.AccessToken)
}

func TestValidateToken(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	password := "ValidateToken123!"
	createReq := &usergrpc.AddRequest{
		Username: "validatetoken_e2e",
		Email:    "validatetoken_e2e@example.com",
		Password: password,
	}

	addResp, err := grpcClient.UserClient.Add(ctx, createReq)
	require.NoError(t, err)

	authReq := &authgrpc.AuthRequest{
		Identifier:        createReq.Email,
		IdentifierType:    "email",
		Password:          password,
		DeviceFingerprint: "validate-device-fingerprint",
		DeviceName:        "Validate Device",
	}

	token, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)

	// When - validate token
	claims, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
		AccessToken: token.AccessToken,
	})

	require.NoError(t, err)
	require.NotNil(t, claims)
	require.Equal(t, addResp.Uid, claims.Uid)
	require.Equal(t, createReq.Email, claims.Identifier)
}

func TestValidateInvalidToken(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	invalidToken := "invalid.token.here"

	// When - validate invalid token
	claims, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
		AccessToken: invalidToken,
	})

	require.Error(t, err)
	require.Nil(t, claims)
}

func TestGetUserProfile(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	createReq := &usergrpc.AddRequest{
		Username: "profileuser_e2e",
		Email:    "profileuser_e2e@example.com",
		Password: "ProfilePassword123!",
	}

	addResp, err := grpcClient.UserClient.Add(ctx, createReq)
	require.NoError(t, err)

	// When - get profile
	profile, err := grpcClient.UserClient.GetProfile(ctx, &usergrpc.GetProfileRequest{
		UserUid: addResp.Uid,
	})

	require.NoError(t, err)
	require.NotNil(t, profile)
}

func TestSetUserPin(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	createReq := &usergrpc.AddRequest{
		Username: "pinuser_e2e",
		Email:    "pinuser_e2e@example.com",
		Password: "PinPassword123!",
	}

	addResp, err := grpcClient.UserClient.Add(ctx, createReq)
	require.NoError(t, err)

	// When - set PIN
	_, err = grpcClient.UserClient.UpdatePin(ctx, &usergrpc.UpdatePinRequest{
		UserUid: addResp.Uid,
		Pin:     "1234",
	})

	require.NoError(t, err)

	// Verify PIN
	verifyResp, err := grpcClient.AuthClient.VerifyPin(ctx, &authgrpc.VerifyPinRequest{
		Uid:  addResp.Uid,
		Code: "1234",
	})
	require.NoError(t, err)
	require.True(t, verifyResp.Valid)
}

func TestVerifyInvalidPin(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	createReq := &usergrpc.AddRequest{
		Username: "invalidpin_e2e",
		Email:    "invalidpin_e2e@example.com",
		Password: "InvalidPin123!",
	}

	addResp, err := grpcClient.UserClient.Add(ctx, createReq)
	require.NoError(t, err)

	// Set PIN
	_, err = grpcClient.UserClient.UpdatePin(ctx, &usergrpc.UpdatePinRequest{
		UserUid: addResp.Uid,
		Pin:     "1234",
	})
	require.NoError(t, err)

	// When - verify with wrong PIN
	verifyResp, err := grpcClient.AuthClient.VerifyPin(ctx, &authgrpc.VerifyPinRequest{
		Uid:  addResp.Uid,
		Code: "5678",
	})

	require.NoError(t, err)
	require.False(t, verifyResp.Valid)
}

func TestUserLoginScenarios(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T) string
		authReq        func(t *testing.T) *authgrpc.AuthRequest
		wantErr        bool
		errContains    string
	}{
		{
			name: "Empty identifier type",
			setup: func(t *testing.T) string {
				return createTestUser(t, nil, "emptytype", "emptytype@example.com", "Password123!")
			},
			authReq: func(t *testing.T) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        "emptytype@example.com",
					IdentifierType:    "",
					Password:          "Password123!",
					DeviceFingerprint: "test-device",
					DeviceName:        "test",
				}
			},
			wantErr: true,
		},
		{
			name: "Inactive user login fails",
			setup: func(t *testing.T) string {
				return createTestUser(t, nil, "inactiveuser", "inactiveuser@example.com", "Password123!")
			},
			authReq: func(t *testing.T) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        "inactiveuser@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: "test-device",
					DeviceName:        "test",
				}
			},
			wantErr:     true,
			errContains: "user account is inactive",
		},
		{
			name: "Login with non-existent user",
			setup: func(t *testing.T) string {
				return "nonexistent"
			},
			authReq: func(t *testing.T) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        "nonexistent@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: "test-device",
					DeviceName:        "test",
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

			_ = tt.setup(t)
			authReq := tt.authReq(t)

			token, err := grpcClient.AuthClient.Auth(ctx, authReq)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, token)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, token)
				require.NotEmpty(t, token.AccessToken)
				require.NotEmpty(t, token.RefreshToken)
			}
		})
	}
}

func TestRefreshTokenTwice(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	createTestUser(t, grpcClient, "refresh2x", "refresh2x@example.com", "Password123!")

	authReq := &authgrpc.AuthRequest{
		Identifier:        "refresh2x@example.com",
		IdentifierType:    "email",
		Password:          "Password123!",
		DeviceFingerprint: "test-device-2x",
		DeviceName:        "Test Device",
	}

	initialToken, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)
	require.NotEmpty(t, initialToken.RefreshToken)

	// When: Refresh token first time
	firstRefresh, err := grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
		RefreshToken: initialToken.RefreshToken,
	})
	require.NoError(t, err)
	require.NotEmpty(t, firstRefresh.RefreshToken)

	// Then: Second refresh should return different token
	secondRefresh, err := grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
		RefreshToken: firstRefresh.RefreshToken,
	})
	require.NoError(t, err)
	require.NotEmpty(t, secondRefresh.RefreshToken)

	// Tokens should be different (token rotation)
	require.NotEqual(t, initialToken.RefreshToken, firstRefresh.RefreshToken)
	require.NotEqual(t, firstRefresh.RefreshToken, secondRefresh.RefreshToken)
}

func TestRefreshTokenExpired(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	createTestUser(t, grpcClient, "expiredtoken", "expiredtoken@example.com", "Password123!")

	authReq := &authgrpc.AuthRequest{
		Identifier:        "expiredtoken@example.com",
		IdentifierType:    "email",
		Password:          "Password123!",
		DeviceFingerprint: "test-device-exp",
		DeviceName:        "Test Device",
	}

	_, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)

	// When: Try to refresh with invalid/expired token
	_, err = grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
		RefreshToken: "invalid.refresh.token",
	})

	// Then: Should fail
	require.Error(t, err)
}

func TestRevokeToken(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	createTestUser(t, grpcClient, "revoketoken", "revoketoken@example.com", "Password123!")

	authReq := &authgrpc.AuthRequest{
		Identifier:        "revoketoken@example.com",
		IdentifierType:    "email",
		Password:          "Password123!",
		DeviceFingerprint: "test-device-revoke",
		DeviceName:        "Test Device",
	}

	token, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)

	// When: Revoke refresh token
	_, err = grpcClient.AuthClient.RevokeToken(ctx, &authgrpc.RevokeTokenRequest{
		Token:     token.RefreshToken,
		TokenType: "refresh",
	})
	require.NoError(t, err)

	// Then: Refresh token should be revoked
	_, err = grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
		RefreshToken: token.RefreshToken,
	})
	require.Error(t, err)
}

func TestLoginWithDeviceTracking(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	uid := createTestUser(t, grpcClient, "devicetrack", "devicetrack@example.com", "Password123!")

	authReq := &authgrpc.AuthRequest{
		Identifier:        "devicetrack@example.com",
		IdentifierType:    "email",
		Password:          "Password123!",
		DeviceFingerprint: "unique-fingerprint-123",
		DeviceName:        "iPhone 14",
	}

	// When: Login for the first time
	token1, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)
	require.NotNil(t, token1)

	// Check that device was created
	devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
		UserUid: uid,
	})
	require.NoError(t, err)
	require.Len(t, devices.Items, 1)
	require.Equal(t, "iPhone 14", devices.Items[0].DeviceName)

	// When: Login again with same device fingerprint
	token2, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)
	require.NotNil(t, token2)

	// Device should be reused (not duplicated)
	devices2, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
		UserUid: uid,
	})
	require.NoError(t, err)
	require.Len(t, devices2.Items, 1) // Still only 1 device
}

func TestListUserDevices(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	uid := createTestUser(t, grpcClient, "multidevice", "multidevice@example.com", "Password123!")

	devices := []struct {
		fingerprint string
		name        string
	}{
		{"fp1", "iPhone 14"},
		{"fp2", "MacBook Pro"},
		{"fp3", "Windows PC"},
	}

	for _, d := range devices {
		authReq := &authgrpc.AuthRequest{
			Identifier:        "multidevice@example.com",
			IdentifierType:    "email",
			Password:          "Password123!",
			DeviceFingerprint: d.fingerprint,
			DeviceName:        d.name,
		}
		_, err := grpcClient.AuthClient.Auth(ctx, authReq)
		require.NoError(t, err)
	}

	// When: List devices
	result, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
		UserUid: uid,
	})

	// Then: All devices should be listed
	require.NoError(t, err)
	require.Len(t, result.Items, 3)

	// Verify device names
	deviceNames := make(map[string]bool)
	for _, d := range result.Items {
		deviceNames[d.DeviceName] = true
	}
	require.True(t, deviceNames["iPhone 14"])
	require.True(t, deviceNames["MacBook Pro"])
	require.True(t, deviceNames["Windows PC"])
}

func TestRevokeDevice(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	uid := createTestUser(t, grpcClient, "revokeuser", "revokeuser@example.com", "Password123!")

	authReq1 := &authgrpc.AuthRequest{
		Identifier:        "revokeuser@example.com",
		IdentifierType:    "email",
		Password:          "Password123!",
		DeviceFingerprint: "fp-keep",
		DeviceName:        "Device to Keep",
	}
	token1, err := grpcClient.AuthClient.Auth(ctx, authReq1)
	require.NoError(t, err)

	authReq2 := &authgrpc.AuthRequest{
		Identifier:        "revokeuser@example.com",
		IdentifierType:    "email",
		Password:          "Password123!",
		DeviceFingerprint: "fp-revoke",
		DeviceName:        "Device to Revoke",
	}
	token2, err := grpcClient.AuthClient.Auth(ctx, authReq2)
	require.NoError(t, err)

	// Get the device UID to revoke
	devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
		UserUid: uid,
	})
	require.NoError(t, err)
	require.Len(t, devices.Items, 2)

	var deviceToRevokeUID string
	for _, d := range devices.Items {
		if d.DeviceName == "Device to Revoke" {
			deviceToRevokeUID = d.DeviceUid
			break
		}
	}
	require.NotEmpty(t, deviceToRevokeUID)

	// When: Revoke device
	_, err = grpcClient.UserClient.RevokeDevice(ctx, &usergrpc.RevokeDeviceRequest{
		UserUid:   uid,
		DeviceUid: deviceToRevokeUID,
	})
	require.NoError(t, err)

	// Then: Device should no longer be listed
	devicesAfter, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
		UserUid: uid,
	})
	require.NoError(t, err)
	require.Len(t, devicesAfter.Items, 1)
	require.Equal(t, "Device to Keep", devicesAfter.Items[0].DeviceName)

	// Token for revoked device should be invalid
	_, err = grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
		AccessToken: token2.AccessToken,
	})
	require.Error(t, err)

	// Token for kept device should still be valid
	claims, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
		AccessToken: token1.AccessToken,
	})
	require.NoError(t, err)
	require.NotNil(t, claims)
}
