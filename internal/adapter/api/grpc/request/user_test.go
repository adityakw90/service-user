package request

import (
	"testing"

	commonpb "github.com/adityakw90/service-user-proto/gen/go/common"
	userpb "github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserGetRequestFromPb(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantUid string
	}{
		{name: "trims surrounding whitespace", raw: "  user-uid  ", wantUid: "user-uid"},
		{name: "already trimmed value unchanged", raw: "user-uid-abc", wantUid: "user-uid-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserGetRequestFromPb(&userpb.GetRequest{Uid: tt.raw})
			assert.Equal(t, tt.wantUid, got.Uid)
		})
	}
}

func TestUserDeleteRequestFromPb(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantUid string
	}{
		{name: "trims surrounding whitespace", raw: "  delete-user-uid  ", wantUid: "delete-user-uid"},
		{name: "already trimmed value unchanged", raw: "delete-user-uid-abc", wantUid: "delete-user-uid-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserDeleteRequestFromPb(&userpb.DeleteRequest{Uid: tt.raw})
			assert.Equal(t, tt.wantUid, got.Uid)
		})
	}
}

func TestUserGetProfileRequestFromPb(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantUserUid string
	}{
		{name: "trims surrounding whitespace", raw: "  profile-user-uid  ", wantUserUid: "profile-user-uid"},
		{name: "already trimmed value unchanged", raw: "profile-user-uid-abc", wantUserUid: "profile-user-uid-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserGetProfileRequestFromPb(&userpb.GetProfileRequest{UserUid: tt.raw})
			assert.Equal(t, tt.wantUserUid, got.UserUid)
		})
	}
}

func TestUserUpdatePinRequestFromPb(t *testing.T) {
	tests := []struct {
		name    string
		rawUid  string
		rawPin  string
		wantUid string
		wantPin string
	}{
		{name: "trims uid and pin", rawUid: "  pin-user-uid  ", rawPin: " 123456 ", wantUid: "pin-user-uid", wantPin: "123456"},
		{name: "already trimmed values unchanged", rawUid: "pin-user-uid-2", rawPin: "654321", wantUid: "pin-user-uid-2", wantPin: "654321"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserUpdatePinRequestFromPb(&userpb.UpdatePinRequest{UserUid: tt.rawUid, Pin: tt.rawPin})
			assert.Equal(t, tt.wantUid, got.UserUid)
			assert.Equal(t, tt.wantPin, got.PIN)
		})
	}
}

func TestUserRevokeDeviceRequestFromPb(t *testing.T) {
	tests := []struct {
		name          string
		rawUserUid    string
		rawDeviceUid  string
		wantUserUid   string
		wantDeviceUid string
	}{
		{name: "trims user and device uids", rawUserUid: "  user-uid  ", rawDeviceUid: "  device-uid  ", wantUserUid: "user-uid", wantDeviceUid: "device-uid"},
		{name: "already trimmed values unchanged", rawUserUid: "user-uid-2", rawDeviceUid: "device-uid-2", wantUserUid: "user-uid-2", wantDeviceUid: "device-uid-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserRevokeDeviceRequestFromPb(&userpb.RevokeDeviceRequest{UserUid: tt.rawUserUid, DeviceUid: tt.rawDeviceUid})
			assert.Equal(t, tt.wantUserUid, got.UserUid)
			assert.Equal(t, tt.wantDeviceUid, got.DeviceUid)
		})
	}
}

func TestUserChangePasswordRequestFromPb(t *testing.T) {
	tests := []struct {
		name        string
		req         *userpb.ChangePasswordRequest
		wantUid     string
		wantCurrent string
		wantNew     string
		wantConfirm string
	}{
		{
			name: "trims all fields",
			req: &userpb.ChangePasswordRequest{
				Uid:             "  user-uid  ",
				CurrentPassword: "  old-pass  ",
				NewPassword:     "  new-pass  ",
				ConfirmPassword: "  new-pass  ",
			},
			wantUid:     "user-uid",
			wantCurrent: "old-pass",
			wantNew:     "new-pass",
			wantConfirm: "new-pass",
		},
		{
			name: "already trimmed values unchanged",
			req: &userpb.ChangePasswordRequest{
				Uid:             "user-uid-2",
				CurrentPassword: "current",
				NewPassword:     "new",
				ConfirmPassword: "new",
			},
			wantUid:     "user-uid-2",
			wantCurrent: "current",
			wantNew:     "new",
			wantConfirm: "new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserChangePasswordRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUid, got.Uid)
			assert.Equal(t, tt.wantCurrent, got.CurrentPassword)
			assert.Equal(t, tt.wantNew, got.NewPassword)
			assert.Equal(t, tt.wantConfirm, got.ConfirmPassword)
		})
	}
}

func TestUserAddRequestFromPb(t *testing.T) {
	tests := []struct {
		name         string
		req          *userpb.AddRequest
		wantUsername string
		wantEmail    string
		wantPassword string
	}{
		{
			name: "trims all fields",
			req: &userpb.AddRequest{
				Username: "  john_doe  ",
				Email:    "  john@example.com  ",
				Password: "  secretpw  ",
			},
			wantUsername: "john_doe",
			wantEmail:    "john@example.com",
			wantPassword: "secretpw",
		},
		{
			name: "already trimmed values unchanged",
			req: &userpb.AddRequest{
				Username: "jane_doe",
				Email:    "jane@example.com",
				Password: "anotherpw",
			},
			wantUsername: "jane_doe",
			wantEmail:    "jane@example.com",
			wantPassword: "anotherpw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserAddRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUsername, got.Username)
			assert.Equal(t, tt.wantEmail, got.Email)
			assert.Equal(t, tt.wantPassword, got.Password)
		})
	}
}

func TestUserUpdateRequestFromPb(t *testing.T) {
	username := " new_username "
	email := " new@example.com "
	password := " new_secret "
	status := int32(1)

	tests := []struct {
		name         string
		req          *userpb.UpdateRequest
		wantUid      string
		wantUsername *string
		wantEmail    *string
		wantPassword *string
		wantStatus   *int32
	}{
		{
			name: "all optional fields populated and trimmed",
			req: &userpb.UpdateRequest{
				Uid:      "  user-uid-123  ",
				Username: &username,
				Email:    &email,
				Password: &password,
				Status:   &status,
			},
			wantUid:      "user-uid-123",
			wantUsername: strPtr("new_username"),
			wantEmail:    strPtr("new@example.com"),
			wantPassword: strPtr("new_secret"),
			wantStatus:   int32Ptr(1),
		},
		{
			name:    "only uid, no optional fields",
			req:     &userpb.UpdateRequest{Uid: "  user-uid-456  "},
			wantUid: "user-uid-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserUpdateRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUid, got.Uid)
			assertOptionalString(t, tt.wantUsername, got.Username, "Username")
			assertOptionalString(t, tt.wantEmail, got.Email, "Email")
			assertOptionalString(t, tt.wantPassword, got.Password, "Password")
			assertOptionalInt32(t, tt.wantStatus, got.StatusPtr, "Status")
		})
	}
}

func TestUserFilterRequestFromPb(t *testing.T) {
	username := " query_user "
	email := " query@example.com "
	status := int32(2)
	exists := true
	query := "search query"

	tests := []struct {
		name       string
		req        *userpb.FilterRequest
		wantUids   []string
		wantUser   *string
		wantEmail  *string
		wantStatus *int32
		wantExists *bool
		wantQuery  *string
	}{
		{
			name: "all filter fields populated",
			req: &userpb.FilterRequest{
				Uids:     []string{"u1", "u2"},
				Username: &username,
				Email:    &email,
				Status:   &status,
				Exists:   &exists,
				Query:    &query,
			},
			wantUids:   []string{"u1", "u2"},
			wantUser:   strPtr("query_user"),
			wantEmail:  strPtr("query@example.com"),
			wantStatus: int32Ptr(2),
			wantExists: boolPtr(true),
			wantQuery:  strPtr("search query"),
		},
		{
			name:     "only Uids set, optional fields unset",
			req:      &userpb.FilterRequest{Uids: []string{"u3"}},
			wantUids: []string{"u3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserFilterRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUids, got.Uids)
			assertOptionalString(t, tt.wantUser, got.Username, "Username")
			assertOptionalString(t, tt.wantEmail, got.Email, "Email")
			assertOptionalInt32(t, tt.wantStatus, got.Status, "Status")
			assertOptionalBool(t, tt.wantExists, got.Exists, "Exists")
			assertOptionalString(t, tt.wantQuery, got.Query, "Query")

			// Verify ToUserFilterParams forwards all fields and converts Status to UserStatus enum.
			params := got.ToUserFilterParams()
			require.NotNil(t, params)
			assert.Equal(t, got.Uids, params.Uids)
			assert.Equal(t, got.Username, params.Username)
			assert.Equal(t, got.Email, params.Email)
			require.NotNil(t, params.Status)
			if tt.wantStatus != nil {
				assert.Equal(t, model.UserStatus(*tt.wantStatus), *params.Status)
			}
			assert.Equal(t, got.Exists, params.Exists)
			assert.Equal(t, got.Query, params.Query)
		})
	}
}

func TestUserUpdateProfileRequestFromPb(t *testing.T) {
	tests := []struct {
		name        string
		req         *userpb.UpdateProfileRequest
		wantUserUid string
		wantFirst   *string
		wantLast    *string
		wantBio     *string
	}{
		{
			name: "all optional fields populated and trimmed",
			req: &userpb.UpdateProfileRequest{
				UserUid:   "  user-uid-abc  ",
				FirstName: "  John  ",
				LastName:  "  Doe  ",
				Bio:       "  Hello world  ",
			},
			wantUserUid: "user-uid-abc",
			wantFirst:   strPtr("John"),
			wantLast:    strPtr("Doe"),
			wantBio:     strPtr("Hello world"),
		},
		{
			name:        "empty strings leave fields nil",
			req:         &userpb.UpdateProfileRequest{UserUid: "  user-uid-abc  "},
			wantUserUid: "user-uid-abc",
		},
		{
			name:        "empty strings leave fields nil (proto-level emptiness check, not whitespace-only)",
			req:         &userpb.UpdateProfileRequest{UserUid: "  user-uid-abc  "},
			wantUserUid: "user-uid-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserUpdateProfileRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUserUid, got.UserUid)
			assertOptionalString(t, tt.wantFirst, got.FirstName, "FirstName")
			assertOptionalString(t, tt.wantLast, got.LastName, "LastName")
			assertOptionalString(t, tt.wantBio, got.Bio, "Bio")
		})
	}
}

func TestUserFilterDeviceRequestFromPb(t *testing.T) {
	name := " My Laptop "
	revoked := true
	emptyName := "   "

	tests := []struct {
		name           string
		req            *userpb.FilterDeviceRequest
		wantDeviceUids []string
		wantDeviceName *string
		wantRevoked    *bool
	}{
		{
			name: "all fields populated and trimmed",
			req: &userpb.FilterDeviceRequest{
				DeviceUids: []string{"d1", "d2"},
				DeviceName: &name,
				Revoked:    &revoked,
			},
			wantDeviceUids: []string{"d1", "d2"},
			wantDeviceName: strPtr("My Laptop"),
			wantRevoked:    boolPtr(true),
		},
		{
			name:           "whitespace-only DeviceName becomes nil",
			req:            &userpb.FilterDeviceRequest{DeviceName: &emptyName},
			wantDeviceName: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserFilterDeviceRequestFromPb(tt.req)
			assert.Equal(t, tt.wantDeviceUids, got.DeviceUids)
			assertOptionalString(t, tt.wantDeviceName, got.DeviceName, "DeviceName")
			assertOptionalBool(t, tt.wantRevoked, got.Revoked, "Revoked")

			// Verify ToUserDeviceListFilterParams forwards fields.
			params := got.ToUserDeviceListFilterParams()
			require.NotNil(t, params)
			assert.Equal(t, got.DeviceUids, params.DeviceUids)
			assert.Equal(t, got.DeviceName, params.DeviceName)
			assert.Equal(t, got.Revoked, params.Revoked)
		})
	}
}

func TestUserListDevicesRequestFromPb(t *testing.T) {
	tests := []struct {
		name           string
		req            *userpb.ListDevicesRequest
		wantUserUid    string
		wantPage       int
		wantLimit      int
		wantSort       string
		wantOrderBy    string
		wantDeviceUids []string
		wantFilterUids []string
	}{
		{
			name: "with pagination and filter forwards supplied fields verbatim",
			req: &userpb.ListDevicesRequest{
				UserUid: "user-uid-1",
				Pagination: &commonpb.Pagination{
					Page:  4,
					Limit: 40,
				},
				Filter: &userpb.FilterDeviceRequest{
					DeviceUids: []string{"d1"},
				},
			},
			wantUserUid:    "user-uid-1",
			wantPage:       4,
			wantLimit:      40,
			wantSort:       "",
			wantOrderBy:    "",
			wantDeviceUids: []string{"d1"},
			wantFilterUids: []string{"user-uid-1"},
		},
		{
			name:           "nil pagination/filter use defaults and inject UserUid into filter",
			req:            &userpb.ListDevicesRequest{UserUid: "user-uid-2"},
			wantUserUid:    "user-uid-2",
			wantPage:       1,
			wantLimit:      10,
			wantSort:       "desc",
			wantOrderBy:    "created_at",
			wantDeviceUids: nil,
			wantFilterUids: []string{"user-uid-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserListDevicesRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUserUid, got.UserUid)

			params := got.ToUserDeviceListParam()
			require.NotNil(t, params)
			require.NotNil(t, params.Pagination)
			assert.Equal(t, tt.wantPage, *params.Pagination.Page)
			assert.Equal(t, tt.wantLimit, *params.Pagination.Limit)
			assert.Equal(t, tt.wantSort, *params.Pagination.Sort)
			assert.Equal(t, tt.wantOrderBy, *params.Pagination.OrderBy)

			require.NotNil(t, params.Filter)
			assert.Equal(t, tt.wantFilterUids, params.Filter.UserUids)
			assert.Equal(t, tt.wantDeviceUids, params.Filter.DeviceUids)
		})
	}
}

func TestUserListRequestFromPb(t *testing.T) {
	tests := []struct {
		name      string
		req       *userpb.ListRequest
		wantPage  int
		wantLimit int
		wantUids  []string
	}{
		{
			name: "with pagination and filter",
			req: &userpb.ListRequest{
				Pagination: &commonpb.Pagination{
					Page:  2,
					Limit: 20,
				},
				Filter: &userpb.FilterRequest{
					Uids: []string{"u1"},
				},
			},
			wantPage:  2,
			wantLimit: 20,
			wantUids:  []string{"u1"},
		},
		{
			name:      "nil pagination/filter apply defaults",
			req:       &userpb.ListRequest{},
			wantPage:  1,
			wantLimit: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserListRequestFromPb(tt.req)
			require.NotNil(t, got)

			params := got.ToUserListParams()
			require.NotNil(t, params)
			require.NotNil(t, params.Pagination)
			assert.Equal(t, tt.wantPage, *params.Pagination.Page)
			assert.Equal(t, tt.wantLimit, *params.Pagination.Limit)
			require.NotNil(t, params.Filter)
			if tt.wantUids != nil {
				assert.Equal(t, tt.wantUids, params.Filter.Uids)
			}
		})
	}
}

// Helper assertions and constructors used across the test files in this package.
// strPtr is defined in auth_test.go.
// assertOptionalString is defined in auth_test.go.

func int32Ptr(v int32) *int32 { return &v }
func boolPtr(v bool) *bool    { return &v }
func assertOptionalInt32(t *testing.T, want, got *int32, field string) {
	t.Helper()
	if want == nil {
		assert.Nil(t, got, field+" should be nil")
	} else {
		require.NotNil(t, got, field+" should not be nil")
		assert.Equal(t, *want, *got, field)
	}
}
func assertOptionalBool(t *testing.T, want, got *bool, field string) {
	t.Helper()
	if want == nil {
		assert.Nil(t, got, field+" should be nil")
	} else {
		require.NotNil(t, got, field+" should not be nil")
		assert.Equal(t, *want, *got, field)
	}
}
