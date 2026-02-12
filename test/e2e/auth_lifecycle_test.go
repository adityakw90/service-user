package e2e

import (
	"context"
	"testing"

	authgrpc "github.com/adityakw90/service-user-proto/gen/go/auth"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	util "github.com/adityakw90/service-user/pkg/util"
	testutil "github.com/adityakw90/service-user/test/util"
	"github.com/stretchr/testify/require"
)

// TestE2E_AuthService_DeviceTracking tests device tracking during authentication
func TestE2E_AuthService_DeviceTracking(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		authReq    func(t *testing.T, email string) *authgrpc.AuthRequest
		verifyFunc func(t *testing.T, uid string, grpcClient *testutil.TestGRPCClient)
	}{
		{
			name: "Login creates new device",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "devicetrack", "devicetrack@example.com", "Password123!")
			},
			authReq: func(t *testing.T, email string) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        email,
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("unique-fingerprint-123"),
					DeviceName:        util.Ptr("iPhone 14"),
				}
			},
			verifyFunc: func(t *testing.T, uid string, grpcClient *testutil.TestGRPCClient) {
				ctx := context.Background()
				devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.Len(t, devices.Items, 1)
				require.Equal(t, "iPhone 14", devices.Items[0].DeviceName)
			},
		},
		{
			name: "Login with same fingerprint reuses device",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				uid := createTestUser(t, grpcClient, "sametrack", "sametrack@example.com", "Password123!")
				ctx := context.Background()
				// First login
				_, _ = grpcClient.AuthClient.Auth(ctx, &authgrpc.AuthRequest{
					Identifier:        "sametrack@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("same-fingerprint"),
					DeviceName:        util.Ptr("Device 1"),
				})
				return uid
			},
			authReq: func(t *testing.T, email string) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        email,
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("same-fingerprint"),
					DeviceName:        util.Ptr("Device 2"),
				}
			},
			verifyFunc: func(t *testing.T, uid string, grpcClient *testutil.TestGRPCClient) {
				ctx := context.Background()
				devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.Len(t, devices.Items, 1) // Still only 1 device
			},
		},
		{
			name: "Multiple devices for same user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "multidevice", "multidevice@example.com", "Password123!")
			},
			authReq: func(t *testing.T, email string) *authgrpc.AuthRequest {
				return &authgrpc.AuthRequest{
					Identifier:        email,
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("fp-3"),
					DeviceName:        util.Ptr("Windows PC"),
				}
			},
			verifyFunc: func(t *testing.T, uid string, grpcClient *testutil.TestGRPCClient) {
				ctx := context.Background()
				// Create multiple devices
				fingerprints := []struct{ fp, name string }{
					{"fp-1", "iPhone 14"},
					{"fp-2", "MacBook Pro"},
				}
				for _, d := range fingerprints {
					_, _ = grpcClient.AuthClient.Auth(ctx, &authgrpc.AuthRequest{
						Identifier:        "multidevice@example.com",
						IdentifierType:    "email",
						Password:          "Password123!",
						DeviceFingerprint: util.Ptr(d.fp),
						DeviceName:        util.Ptr(d.name),
					})
				}
				devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.Len(t, devices.Items, 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			uid := tt.setup(t, grpcClient)
			authReq := tt.authReq(t, "multidevice@example.com")
			if tt.name == "Login creates new device" {
				authReq = tt.authReq(t, "devicetrack@example.com")
			} else if tt.name == "Login with same fingerprint reuses device" {
				authReq = tt.authReq(t, "sametrack@example.com")
			}

			_, err := grpcClient.AuthClient.Auth(ctx, authReq)
			require.NoError(t, err)

			if tt.verifyFunc != nil {
				tt.verifyFunc(t, uid, grpcClient)
			}
		})
	}
}

// TestE2E_AuthService_RevokeDevice tests device revocation via user service
func TestE2E_AuthService_RevokeDevice(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string)
		wantErr    bool
		verifyFunc func(t *testing.T, uid string, grpcClient *testutil.TestGRPCClient, tokens map[string]*authgrpc.Token)
	}{
		{
			name: "Revoke device and verify token invalidation",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string) {
				uid := createTestUser(t, grpcClient, "revokeuser", "revokeuser@example.com", "Password123!")
				ctx := context.Background()

				// Create two devices
				authReq1 := &authgrpc.AuthRequest{
					Identifier:        "revokeuser@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("fp-keep"),
					DeviceName:        util.Ptr("Device to Keep"),
				}
				_, _ = grpcClient.AuthClient.Auth(ctx, authReq1)

				authReq2 := &authgrpc.AuthRequest{
					Identifier:        "revokeuser@example.com",
					IdentifierType:    "email",
					Password:          "Password123!",
					DeviceFingerprint: util.Ptr("fp-revoke"),
					DeviceName:        util.Ptr("Device to Revoke"),
				}
				_, _ = grpcClient.AuthClient.Auth(ctx, authReq2)

				// Get the device UID to revoke
				devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)

				var deviceToRevokeUID string
				for _, d := range devices.Items {
					if d.DeviceName == "Device to Revoke" {
						deviceToRevokeUID = d.DeviceUid
						break
					}
				}
				require.NotEmpty(t, deviceToRevokeUID)

				return uid, deviceToRevokeUID
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, uid string, grpcClient *testutil.TestGRPCClient, tokens map[string]*authgrpc.Token) {
				ctx := context.Background()
				// Device should no longer be listed
				devicesAfter, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.Len(t, devicesAfter.Items, 1)
				require.Equal(t, "Device to Keep", devicesAfter.Items[0].DeviceName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			uid, deviceUID := tt.setup(t, grpcClient)

			_, err := grpcClient.UserClient.RevokeDevice(ctx, &usergrpc.RevokeDeviceRequest{
				UserUid:   uid,
				DeviceUid: deviceUID,
			})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, uid, grpcClient, nil)
				}
			}
		})
	}
}
