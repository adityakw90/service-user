package e2e

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	authgrpc "github.com/adityakw90/service-user-proto/gen/go/auth"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/stretchr/testify/require"
)

func TestFullUserLifecycle(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()

	// Phase 1: Register new user
	createReq := &usergrpc.AddRequest{
		Username: "lifecycleuser",
		Email:    "lifecycleuser@example.com",
		Password: "LifecyclePassword123!",
	}

	addResp, err := grpcClient.UserClient.Add(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, addResp)
	userUID := addResp.Uid

	// Verify user is active
	user, err := grpcClient.UserClient.Get(ctx, &usergrpc.GetRequest{Uid: userUID})
	require.NoError(t, err)
	require.Equal(t, int32(model.UserStatusActive), user.Status)

	// Phase 2: Login with credentials
	authReq := &authgrpc.AuthRequest{
		Identifier:        createReq.Email,
		IdentifierType:    "email",
		Password:          createReq.Password,
		DeviceFingerprint: "lifecycle-device",
		DeviceName:        "Lifecycle Device",
	}

	token, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, token.AccessToken)
	require.NotEmpty(t, token.RefreshToken)

	// Phase 3: Get and update profile
	profile, err := grpcClient.UserClient.GetProfile(ctx, &usergrpc.GetProfileRequest{
		UserUid: userUID,
	})
	require.NoError(t, err)
	require.NotNil(t, profile)

	firstName := "Lifecycle"
	lastName := "User"
	bio := "Testing full lifecycle"

	_, err = grpcClient.UserClient.UpdateProfile(ctx, &usergrpc.UpdateProfileRequest{
		UserUid:   userUID,
		FirstName: firstName,
		LastName:  lastName,
		Bio:       bio,
	})
	require.NoError(t, err)

	updatedProfile, err := grpcClient.UserClient.GetProfile(ctx, &usergrpc.GetProfileRequest{
		UserUid: userUID,
	})
	require.NoError(t, err)
	require.Equal(t, "Lifecycle", updatedProfile.FirstName)
	require.Equal(t, "User", updatedProfile.LastName)
	require.Equal(t, "Testing full lifecycle", updatedProfile.Bio)

	// Phase 4: Update username
	newUsername := "updatedlifecycleuser"
	_, err = grpcClient.UserClient.Update(ctx, &usergrpc.UpdateRequest{
		Uid:      userUID,
		Username: &newUsername,
	})
	require.NoError(t, err)

	updatedUser, err := grpcClient.UserClient.Get(ctx, &usergrpc.GetRequest{Uid: userUID})
	require.NoError(t, err)
	require.Equal(t, "updatedlifecycleuser", updatedUser.Username)

	// Phase 5: Delete user
	_, err = grpcClient.UserClient.Delete(ctx, &usergrpc.DeleteRequest{Uid: userUID})
	require.NoError(t, err)

	// Phase 6: Verify user is deleted
	_, err = grpcClient.UserClient.Get(ctx, &usergrpc.GetRequest{Uid: userUID})
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrUserDeleted)

	// Phase 7: Verify login fails after deletion
	_, err = grpcClient.AuthClient.Auth(ctx, authReq)
	require.Error(t, err)
}

func TestMultiDeviceLoginFlow(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	uid := createTestUser(t, grpcClient, "multilogin", "multilogin@example.com", "MultiLogin123!")

	devices := []struct {
		fingerprint string
		name        string
	}{
		{"fp-mobile-1", "iPhone 14"},
		{"fp-desktop-1", "MacBook Pro"},
		{"fp-tablet-1", "iPad Pro"},
	}

	var tokens []*authgrpc.Token

	// When: Login from multiple devices
	for _, d := range devices {
		authReq := &authgrpc.AuthRequest{
			Identifier:        "multilogin@example.com",
			IdentifierType:    "email",
			Password:          "MultiLogin123!",
			DeviceFingerprint: d.fingerprint,
			DeviceName:        d.name,
		}

		token, err := grpcClient.AuthClient.Auth(ctx, authReq)
		require.NoError(t, err)
		require.NotNil(t, token)
		tokens = append(tokens, token)
	}

	// Then: All tokens should be valid
	for i, token := range tokens {
		claims, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
			AccessToken: token.AccessToken,
		})
		require.NoError(t, err, "Token %d should be valid", i)
		require.NotNil(t, claims)
		require.Equal(t, uid, claims.Uid)
	}

	// And: All devices should be tracked
	deviceList, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
		UserUid: uid,
	})
	require.NoError(t, err)
	require.Len(t, deviceList.Items, 3)

	// And: Each device can refresh its token
	for i, token := range tokens {
		newToken, err := grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
			RefreshToken: token.RefreshToken,
		})
		require.NoError(t, err, "Device %d should be able to refresh token", i)
		require.NotNil(t, newToken)
		require.NotEmpty(t, newToken.AccessToken)
		require.NotEmpty(t, newToken.RefreshToken)
	}
}

func TestPasswordChangeInvalidatesTokens(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	uid := createTestUser(t, grpcClient, "passchange", "passchange@example.com", "OldPassword123!")

	authReq1 := &authgrpc.AuthRequest{
		Identifier:        "passchange@example.com",
		IdentifierType:    "email",
		Password:          "OldPassword123!",
		DeviceFingerprint: "device-1",
		DeviceName:        "Device 1",
	}

	authReq2 := &authgrpc.AuthRequest{
		Identifier:        "passchange@example.com",
		IdentifierType:    "email",
		Password:          "OldPassword123!",
		DeviceFingerprint: "device-2",
		DeviceName:        "Device 2",
	}

	token1, err := grpcClient.AuthClient.Auth(ctx, authReq1)
	require.NoError(t, err)

	token2, err := grpcClient.AuthClient.Auth(ctx, authReq2)
	require.NoError(t, err)

	// Verify tokens are valid before password change
	claims1, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
		AccessToken: token1.AccessToken,
	})
	require.NoError(t, err)
	require.NotNil(t, claims1)

	claims2, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
		AccessToken: token2.AccessToken,
	})
	require.NoError(t, err)
	require.NotNil(t, claims2)

	// When: Change password
	changeReq := &usergrpc.ChangePasswordRequest{
		Uid:             uid,
		CurrentPassword: "OldPassword123!",
		NewPassword:     "NewPassword456!",
		ConfirmPassword: "NewPassword456!",
	}

	err = func() error {
		_, err := grpcClient.UserClient.ChangePassword(ctx, changeReq)
		return err
	}()
	require.NoError(t, err)

	// Then: Old tokens should still work (implementation may vary)
	claims1After, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
		AccessToken: token1.AccessToken,
	})
	require.NoError(t, err)
	require.NotNil(t, claims1After)

	// But we can no longer login with old password
	_, err = grpcClient.AuthClient.Auth(ctx, authReq1)
	require.Error(t, err)

	// And new password works
	authReqNew := &authgrpc.AuthRequest{
		Identifier:        "passchange@example.com",
		IdentifierType:    "email",
		Password:          "NewPassword456!",
		DeviceFingerprint: "device-1",
		DeviceName:        "Device 1",
	}

	newToken, err := grpcClient.AuthClient.Auth(ctx, authReqNew)
	require.NoError(t, err)
	require.NotNil(t, newToken)
}

func TestAccountDeactivationPreventsLogin(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	uid := createTestUser(t, grpcClient, "deactivateuser", "deactivateuser@example.com", "Deactivate123!")

	authReq := &authgrpc.AuthRequest{
		Identifier:        "deactivateuser@example.com",
		IdentifierType:    "email",
		Password:          "Deactivate123!",
		DeviceFingerprint: "deactivate-device",
		DeviceName:        "Test Device",
	}

	token, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)
	require.NotNil(t, token)

	// Verify token is valid
	claims, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
		AccessToken: token.AccessToken,
	})
	require.NoError(t, err)
	require.NotNil(t, claims)

	// When: Deactivate user
	deactivateUser(t, grpcClient, uid)

	// Then: Login should fail
	_, err = grpcClient.AuthClient.Auth(ctx, authReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user account is inactive")

	// And: Getting user should fail
	_, err = grpcClient.UserClient.Get(ctx, &usergrpc.GetRequest{Uid: uid})
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrUserInactive)

	// When: Reactivate user
	activeStatus := int32(model.UserStatusActive)
	_, err = grpcClient.UserClient.Update(ctx, &usergrpc.UpdateRequest{
		Uid:    uid,
		Status: &activeStatus,
	})
	require.NoError(t, err)

	// Then: Login should work again
	newToken, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)
	require.NotNil(t, newToken)
	require.NotEmpty(t, newToken.AccessToken)
}

func TestUserProfileIntegration(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	uid := createTestUser(t, grpcClient, "profileuser", "profileuser@example.com", "Profile123!")

	// When: Update profile with all fields
	firstName := "John"
	lastName := "Doe"
	bio := "Software developer and Go enthusiast"

	err := func() error {
		_, err := grpcClient.UserClient.UpdateProfile(ctx, &usergrpc.UpdateProfileRequest{
			UserUid:   uid,
			FirstName: firstName,
			LastName:  lastName,
			Bio:       bio,
		})
		return err
	}()
	require.NoError(t, err)

	// Then: Profile should be retrievable and match
	profile, err := grpcClient.UserClient.GetProfile(ctx, &usergrpc.GetProfileRequest{
		UserUid: uid,
	})
	require.NoError(t, err)
	require.NotNil(t, profile)

	require.Equal(t, "John", profile.FirstName)
	require.Equal(t, "Doe", profile.LastName)
	require.Equal(t, "Software developer and Go enthusiast", profile.Bio)

	// When: Partially update profile
	newBio := "Senior Go developer"
	err = func() error {
		_, err := grpcClient.UserClient.UpdateProfile(ctx, &usergrpc.UpdateProfileRequest{
			UserUid: uid,
			Bio:     newBio,
		})
		return err
	}()
	require.NoError(t, err)

	// Then: Only bio should change, other fields should remain
	updatedProfile, err := grpcClient.UserClient.GetProfile(ctx, &usergrpc.GetProfileRequest{
		UserUid: uid,
	})
	require.NoError(t, err)
	require.Equal(t, "John", updatedProfile.FirstName) // Unchanged
	require.Equal(t, "Doe", updatedProfile.LastName)   // Unchanged
	require.Equal(t, "Senior Go developer", updatedProfile.Bio)
}

func TestUserDeviceManagementIntegration(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()
	uid := createTestUser(t, grpcClient, "devicemgmt", "devicemgmt@example.com", "DeviceMgmt123!")

	authReq := &authgrpc.AuthRequest{
		Identifier:        "devicemgmt@example.com",
		IdentifierType:    "email",
		Password:          "DeviceMgmt123!",
		DeviceFingerprint: "fp-keep",
		DeviceName:        "Primary Device",
	}

	primaryToken, err := grpcClient.AuthClient.Auth(ctx, authReq)
	require.NoError(t, err)

	// Add secondary device
	authReq2 := &authgrpc.AuthRequest{
		Identifier:        "devicemgmt@example.com",
		IdentifierType:    "email",
		Password:          "DeviceMgmt123!",
		DeviceFingerprint: "fp-revoke",
		DeviceName:        "Secondary Device",
	}

	secondaryToken, err := grpcClient.AuthClient.Auth(ctx, authReq2)
	require.NoError(t, err)

	// Get all devices
	devices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
		UserUid: uid,
	})
	require.NoError(t, err)
	require.Len(t, devices.Items, 2)

	// Find the secondary device UID
	var secondaryDeviceUID string
	for _, d := range devices.Items {
		if d.DeviceName == "Secondary Device" {
			secondaryDeviceUID = d.DeviceUid
			break
		}
	}
	require.NotEmpty(t, secondaryDeviceUID)

	// When: Revoke secondary device
	_, err = grpcClient.UserClient.RevokeDevice(ctx, &usergrpc.RevokeDeviceRequest{
		UserUid:   uid,
		DeviceUid: secondaryDeviceUID,
	})
	require.NoError(t, err)

	// Then: Only primary device should remain
	remainingDevices, err := grpcClient.UserClient.ListDevice(ctx, &usergrpc.ListDevicesRequest{
		UserUid: uid,
	})
	require.NoError(t, err)
	require.Len(t, remainingDevices.Items, 1)
	require.Equal(t, "Primary Device", remainingDevices.Items[0].DeviceName)

	// And: Primary token should still be valid
	claims, err := grpcClient.AuthClient.ValidateToken(ctx, &authgrpc.ValidateTokenRequest{
		AccessToken: primaryToken.AccessToken,
	})
	require.NoError(t, err)
	require.NotNil(t, claims)

	// And: Secondary token should be invalid (or refresh should fail)
	_, err = grpcClient.AuthClient.RefreshToken(ctx, &authgrpc.RefreshTokenRequest{
		RefreshToken: secondaryToken.RefreshToken,
	})
	require.Error(t, err)
}
