package e2e

import (
	"context"
	"testing"

	authgrpc "github.com/adityakw90/service-user-proto/gen/go/auth"
	commongrpc "github.com/adityakw90/service-user-proto/gen/go/common"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/pkg/util"
	testutil "github.com/adityakw90/service-user/test/util"
	"github.com/stretchr/testify/require"
)

func TestUserGet(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name    string
		setup   func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		wantErr bool
		errMsg  string
	}{
		{
			name: "Get existing user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "getuser", "getuser@example.com", "Password123!")
			},
			wantErr: false,
		},
		{
			name: "Get non-existent user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "01234567-89ab-cdef-0123-456789abcdef"
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name: "Get deleted user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				uid := createTestUser(t, grpcClient, "deleteduser", "deleteduser@example.com", "Password123!")
				deleteUser(t, grpcClient, uid)
				return uid
			},
			wantErr: true,
			errMsg:  "user has been deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			uid := tt.setup(t, grpcClient)

			user, err := grpcClient.UserClient.Get(ctx, &usergrpc.GetRequest{Uid: uid})

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, user)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, uid, user.Uid)
			}
		})
	}
}

func TestUserList(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()

	uid1 := createTestUser(t, grpcClient, "listuser1", "listuser1@example.com", "Password123!")
	uid2 := createTestUser(t, grpcClient, "listuser2", "listuser2@example.com", "Password123!")
	uid3 := createTestUser(t, grpcClient, "listuser3", "listuser3@example.com", "Password123!")

	tests := []struct {
		name       string
		pagination *commongrpc.Pagination
		filter     *usergrpc.FilterRequest
		wantCount  int
		wantUIDs   []string
	}{
		{
			name: "List all users",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   10,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter:    nil,
			wantCount: 3,
			wantUIDs:  []string{uid1, uid2, uid3},
		},
		{
			name: "List with pagination",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   2,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter:    nil,
			wantCount: 2,
			wantUIDs:  []string{uid1, uid2},
		},
		{
			name: "List filtered by email",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   10,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter: &usergrpc.FilterRequest{
				Email: util.Ptr("listuser2@example.com"),
			},
			wantCount: 1,
			wantUIDs:  []string{uid2},
		},
		{
			name: "List filtered by username",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   10,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter: &usergrpc.FilterRequest{
				Username: util.Ptr("listuser1"),
			},
			wantCount: 1,
			wantUIDs:  []string{uid1},
		},
		{
			name: "List filtered by UIDs",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   10,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter: &usergrpc.FilterRequest{
				Uids: []string{uid1, uid3},
			},
			wantCount: 2,
			wantUIDs:  []string{uid1, uid3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := grpcClient.UserClient.List(ctx, &usergrpc.ListRequest{
				Pagination: tt.pagination,
				Filter:     tt.filter,
			})

			require.NoError(t, err)
			require.NotNil(t, users)
			require.Len(t, users.Items, tt.wantCount)

			uidMap := make(map[string]bool)
			for _, user := range users.Items {
				uidMap[user.Uid] = true
			}
			for _, uid := range tt.wantUIDs {
				require.True(t, uidMap[uid], "UID %s should be in results", uid)
			}
		})
	}
}

func TestUserUpdate(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name           string
		setup          func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		update         func(t *testing.T) *usergrpc.UpdateRequest
		wantErr        bool
		verifyGetFails bool
		verifyFunc     func(t *testing.T, user *usergrpc.User)
	}{
		{
			name: "Update username",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "updateuser", "updateuser@example.com", "Password123!")
			},
			update: func(t *testing.T) *usergrpc.UpdateRequest {
				newUsername := "updatedusername"
				return &usergrpc.UpdateRequest{
					Username: &newUsername,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, user *usergrpc.User) {
				require.Equal(t, "updatedusername", user.Username)
			},
		},
		{
			name: "Update email",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "updateemail", "updateemail@example.com", "Password123!")
			},
			update: func(t *testing.T) *usergrpc.UpdateRequest {
				newEmail := "updatedemail@example.com"
				return &usergrpc.UpdateRequest{
					Email: &newEmail,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, user *usergrpc.User) {
				require.Equal(t, "updatedemail@example.com", user.Email)
			},
		},
		{
			name: "Update password",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "updatepass", "updatepass@example.com", "Password123!")
			},
			update: func(t *testing.T) *usergrpc.UpdateRequest {
				newPassword := "NewPassword456!"
				return &usergrpc.UpdateRequest{
					Password: &newPassword,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, user *usergrpc.User) {
				require.NotNil(t, user)
			},
		},
		{
			name: "Update status to inactive",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "updatestatus", "updatestatus@example.com", "Password123!")
			},
			update: func(t *testing.T) *usergrpc.UpdateRequest {
				status := int32(0) // Inactive
				return &usergrpc.UpdateRequest{
					Status: &status,
				}
			},
			wantErr: false,
			// After deactivating, getting the user should fail
			verifyGetFails: true,
		},
		{
			name: "Update multiple fields",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "updatemulti", "updatemulti@example.com", "Password123!")
			},
			update: func(t *testing.T) *usergrpc.UpdateRequest {
				newUsername := "multiupdated"
				newEmail := "multiupdated@example.com"
				return &usergrpc.UpdateRequest{
					Username: &newUsername,
					Email:    &newEmail,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, user *usergrpc.User) {
				require.Equal(t, "multiupdated", user.Username)
				require.Equal(t, "multiupdated@example.com", user.Email)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			uid := tt.setup(t, grpcClient)
			updateReq := tt.update(t)
			updateReq.Uid = uid

			err := func() error {
				_, err := grpcClient.UserClient.Update(ctx, updateReq)
				return err
			}()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.verifyGetFails {
					// After updating to inactive, getting the user should fail
					_, err := grpcClient.UserClient.Get(ctx, &usergrpc.GetRequest{Uid: uid})
					require.Error(t, err)
					require.Contains(t, err.Error(), "user account is inactive")
				} else {
					user, err := grpcClient.UserClient.Get(ctx, &usergrpc.GetRequest{Uid: uid})
					require.NoError(t, err)
					require.NotNil(t, user)

					if tt.verifyFunc != nil {
						tt.verifyFunc(t, user)
					}
				}
			}
		})
	}
}

func TestUserDelete(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name    string
		setup   func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		wantErr bool
	}{
		{
			name: "Soft delete existing user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "deleteuser", "deleteuser@example.com", "Password123!")
			},
			wantErr: false,
		},
		{
			name: "Delete non-existent user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "non-existent-uid"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			uid := tt.setup(t, grpcClient)

			err := func() error {
				_, err := grpcClient.UserClient.Delete(ctx, &usergrpc.DeleteRequest{Uid: uid})
				return err
			}()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				user, err := grpcClient.UserClient.Get(ctx, &usergrpc.GetRequest{Uid: uid})
				require.Error(t, err)
				require.Nil(t, user)
			}
		})
	}
}

func TestChangePassword(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name             string
		setup            func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		changePasswordFn func(t *testing.T) *usergrpc.ChangePasswordRequest
		wantErr          bool
		errMsg           string
	}{
		{
			name: "Change password with correct current password",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "changepass", "changepass@example.com", "CurrentPassword123!")
			},
			changePasswordFn: func(t *testing.T) *usergrpc.ChangePasswordRequest {
				return &usergrpc.ChangePasswordRequest{
					CurrentPassword: "CurrentPassword123!",
					NewPassword:     "NewPassword456!",
					ConfirmPassword: "NewPassword456!",
				}
			},
			wantErr: false,
		},
		{
			name: "Change password with wrong current password",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "wrongpass", "wrongpass@example.com", "CorrectPassword123!")
			},
			changePasswordFn: func(t *testing.T) *usergrpc.ChangePasswordRequest {
				return &usergrpc.ChangePasswordRequest{
					CurrentPassword: "WrongPassword123!",
					NewPassword:     "NewPassword456!",
					ConfirmPassword: "NewPassword456!",
				}
			},
			wantErr: true,
			errMsg:  "current password is incorrect",
		},
		{
			name: "Change password for non-existent user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "01234567-89ab-cdef-0123-456789abcdef"
			},
			changePasswordFn: func(t *testing.T) *usergrpc.ChangePasswordRequest {
				return &usergrpc.ChangePasswordRequest{
					CurrentPassword: "AnyPassword123!",
					NewPassword:     "NewPassword456!",
					ConfirmPassword: "NewPassword456!",
				}
			},
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			uid := tt.setup(t, grpcClient)
			changeReq := tt.changePasswordFn(t)
			changeReq.Uid = uid

			err := func() error {
				_, err := grpcClient.UserClient.ChangePassword(ctx, changeReq)
				return err
			}()

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)

				authReq := &authgrpc.AuthRequest{
					Identifier:        "changepass@example.com",
					IdentifierType:    "email",
					Password:          "NewPassword456!",
					DeviceFingerprint: util.Ptr("test-device"),
					DeviceName:        util.Ptr("test"),
				}

				token, err := grpcClient.AuthClient.Auth(ctx, authReq)
				require.NoError(t, err)
				require.NotNil(t, token)
				require.NotEmpty(t, token.AccessToken)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		updateFn   func(t *testing.T) *usergrpc.UpdateProfileRequest
		wantErr    bool
		verifyFunc func(t *testing.T, profile *usergrpc.Profile)
	}{
		{
			name: "Update all profile fields",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "profilefull", "profilefull@example.com", "Password123!")
			},
			updateFn: func(t *testing.T) *usergrpc.UpdateProfileRequest {
				firstName := "John"
				lastName := "Doe"
				bio := "Software Engineer"
				return &usergrpc.UpdateProfileRequest{
					FirstName: firstName,
					LastName:  lastName,
					Bio:       bio,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, profile *usergrpc.Profile) {
				require.Equal(t, "John", profile.FirstName)
				require.Equal(t, "Doe", profile.LastName)
				require.Equal(t, "Software Engineer", profile.Bio)
			},
		},
		{
			name: "Partial profile update",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "profilepartial", "profilepartial@example.com", "Password123!")
			},
			updateFn: func(t *testing.T) *usergrpc.UpdateProfileRequest {
				firstName := "Jane"
				return &usergrpc.UpdateProfileRequest{
					FirstName: firstName,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, profile *usergrpc.Profile) {
				require.Equal(t, "Jane", profile.FirstName)
				require.Empty(t, profile.LastName)
				require.Empty(t, profile.Bio)
			},
		},
		{
			name: "Update profile for non-existent user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "non-existent-uid"
			},
			updateFn: func(t *testing.T) *usergrpc.UpdateProfileRequest {
				firstName := "Ghost"
				return &usergrpc.UpdateProfileRequest{
					FirstName: firstName,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			uid := tt.setup(t, grpcClient)
			updateReq := tt.updateFn(t)
			updateReq.UserUid = uid

			err := func() error {
				_, err := grpcClient.UserClient.UpdateProfile(ctx, updateReq)
				return err
			}()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				profile, err := grpcClient.UserClient.GetProfile(ctx, &usergrpc.GetProfileRequest{
					UserUid: uid,
				})
				require.NoError(t, err)
				require.NotNil(t, profile)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, profile)
				}
			}
		})
	}
}
