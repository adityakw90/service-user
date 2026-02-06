package e2e

import (
	"context"
	"testing"

	authgrpc "github.com/adityakw90/service-user-proto/gen/go/auth"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/test/util"
	"github.com/stretchr/testify/require"
)

// TestE2E_DeviceService_Lifecycle_LoginScenario tests the device lifecycle through login flow
func TestE2E_DeviceService_Lifecycle_LoginScenario(t *testing.T) {
	tests := []struct {
		name     string
		scenario func(t *testing.T, grpcClient *util.TestGRPCClient, uid string)
		setup    func(t *testing.T, grpcClient *util.TestGRPCClient) string
	}{
		{
			name: "Device created on first login",
			setup: func(t *testing.T, grpcClient *util.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "lifecycle1", "lifecycle1@example.com", "Password123!")
			},
			scenario: func(t *testing.T, grpcClient *util.TestGRPCClient, uid string) {
				ctx := context.Background()

				// First login - creates device
				authReq := &authgrpc.AuthRequest{
					Identifier:        "lifecycle1@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: "lifecycle-fp-1",
					DeviceName:        "First Device",
				}
				token, err := grpcClient.AuthClient.Auth(ctx, authReq)
				require.NoError(t, err)
				require.NotNil(t, token)

				// Verify device was created
				devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.Len(t, devices.Items, 1)
				require.Equal(t, "First Device", devices.Items[0].DeviceName)
			},
		},
		{
			name: "Device reused on subsequent login with same fingerprint",
			setup: func(t *testing.T, grpcClient *util.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "lifecycle2", "lifecycle2@example.com", "Password123!")
			},
			scenario: func(t *testing.T, grpcClient *util.TestGRPCClient, uid string) {
				ctx := context.Background()

				// First login
				authReq1 := &authgrpc.AuthRequest{
					Identifier:        "lifecycle2@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: "lifecycle-fp-2",
					DeviceName:        "Device A",
				}
				token1, err := grpcClient.AuthClient.Auth(ctx, authReq1)
				require.NoError(t, err)
				require.NotNil(t, token1)

				// Second login with same fingerprint but different name
				authReq2 := &authgrpc.AuthRequest{
					Identifier:        "lifecycle2@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: "lifecycle-fp-2", // Same fingerprint
					DeviceName:        "Device B",       // Different name
				}
				token2, err := grpcClient.AuthClient.Auth(ctx, authReq2)
				require.NoError(t, err)
				require.NotNil(t, token2)

				// Verify only one device exists
				devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.Len(t, devices.Items, 1) // Still only 1 device
			},
		},
		{
			name: "Multiple devices created with different fingerprints",
			setup: func(t *testing.T, grpcClient *util.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "lifecycle3", "lifecycle3@example.com", "Password123!")
			},
			scenario: func(t *testing.T, grpcClient *util.TestGRPCClient, uid string) {
				ctx := context.Background()

				// Login from multiple devices
				fingerprints := []struct {
					fingerprint string
					name        string
				}{
					{"lifecycle-fp-3a", "Mobile Phone"},
					{"lifecycle-fp-3b", "Laptop"},
					{"lifecycle-fp-3c", "Tablet"},
				}

				for _, fp := range fingerprints {
					authReq := &authgrpc.AuthRequest{
						Identifier:        "lifecycle3@example.com",
						IdentifierType:    "email",
						Password:          "Password123!",
						DeviceFingerprint: fp.fingerprint,
						DeviceName:        fp.name,
					}
					_, err := grpcClient.AuthClient.Auth(ctx, authReq)
					require.NoError(t, err)
				}

				// Verify all three devices exist
				devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.Len(t, devices.Items, 3)
			},
		},
		{
			name: "Device revocation lifecycle",
			setup: func(t *testing.T, grpcClient *util.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "lifecycle4", "lifecycle4@example.com", "Password123!")
			},
			scenario: func(t *testing.T, grpcClient *util.TestGRPCClient, uid string) {
				ctx := context.Background()

				// Create two devices
				authReq1 := &authgrpc.AuthRequest{
					Identifier:        "lifecycle4@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: "lifecycle-fp-4a",
					DeviceName:        "Device to Keep",
				}
				tokenKeep, err := grpcClient.AuthClient.Auth(ctx, authReq1)
				require.NoError(t, err)

				authReq2 := &authgrpc.AuthRequest{
					Identifier:        "lifecycle4@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: "lifecycle-fp-4b",
					DeviceName:        "Device to Revoke",
				}
				tokenRevoke, err := grpcClient.AuthClient.Auth(ctx, authReq2)
				require.NoError(t, err)

				// Get device UID to revoke
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

				// Revoke the device
				_, err = grpcClient.UserClient.RevokeDevice(ctx, &usergrpc.RevokeDeviceRequest{
					UserUid:   uid,
					DeviceUid: deviceToRevokeUID,
				})
				require.NoError(t, err)

				// Verify device was revoked (no longer in list)
				devicesAfter, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.Len(t, devicesAfter.Items, 1)
				require.Equal(t, "Device to Keep", devicesAfter.Items[0].DeviceName)

				// Verify token for revoked device is invalid
				_, err = grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
					AccessToken: tokenRevoke.AccessToken,
				})
				require.Error(t, err)

				// Verify token for kept device is still valid
				claims, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
					AccessToken: tokenKeep.AccessToken,
				})
				require.NoError(t, err)
				require.NotNil(t, claims)
			},
		},
		{
			name: "Token refresh maintains device association",
			setup: func(t *testing.T, grpcClient *util.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "lifecycle5", "lifecycle5@example.com", "Password123!")
			},
			scenario: func(t *testing.T, grpcClient *util.TestGRPCClient, uid string) {
				ctx := context.Background()

				// Initial login
				authReq := &authgrpc.AuthRequest{
					Identifier:        "lifecycle5@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: "lifecycle-fp-5",
					DeviceName:        "Refresh Test Device",
				}
				initialToken, err := grpcClient.AuthClient.Auth(ctx, authReq)
				require.NoError(t, err)

				// Refresh token
				newToken, err := grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
					RefreshToken: initialToken.RefreshToken,
				})
				require.NoError(t, err)

				// Verify both tokens are valid
				claims1, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
					AccessToken: initialToken.AccessToken,
				})
				require.NoError(t, err)
				require.NotNil(t, claims1)

				claims2, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
					AccessToken: newToken.AccessToken,
				})
				require.NoError(t, err)
				require.NotNil(t, claims2)

				// Verify still only one device
				devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.Len(t, devices.Items, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, grpcClient, cleanup := setupE2ETest(t)
			defer cleanup()

			uid := tt.setup(t, grpcClient)
			tt.scenario(t, grpcClient, uid)
		})
	}
}
