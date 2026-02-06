package fixture

import (
	"fmt"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
)

// ValidTestUser returns a valid UserCreateParam for testing.
// The user has valid username, email, and password that meet all validation requirements.
func ValidTestUser() *params.UserCreateParam {
	return &params.UserCreateParam{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "SecurePassword123!",
	}
}

// RandomTestUser generates a user with random data for testing.
// Use this when you need to create multiple users in tests without conflicts.
func RandomTestUser(suffix int) *params.UserCreateParam {
	return &params.UserCreateParam{
		Username: fmt.Sprintf("randomuser_%d", suffix),
		Email:    fmt.Sprintf("randomuser_%d@example.com", suffix),
		Password: "RandomPassword123!",
	}
}

// ValidUserModel returns a valid User model entity.
// The password should already be hashed before setting.
func ValidUserModel() *model.User {
	return &model.User{
		UID:      "valid-test-user-uid",
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "", // Should be set with hashed password
		Status:   model.UserStatusActive,
	}
}

// InactiveUserModel returns a User with inactive status.
func InactiveUserModel() *model.User {
	user := ValidUserModel()
	user.UID = "inactive-user-uid"
	user.Username = "inactiveuser"
	user.Email = "inactive@example.com"
	user.Status = model.UserStatusInactive
	return user
}

// BannedUserModel returns a User with banned status.
func BannedUserModel() *model.User {
	user := ValidUserModel()
	user.UID = "banned-user-uid"
	user.Username = "banneduser"
	user.Email = "banned@example.com"
	user.Status = model.UserStatusBanned
	return user
}

// ValidUserProfile returns a valid UserProfile model entity.
func ValidUserProfile() *model.UserProfile {
	return &model.UserProfile{
		FirstName:  "Test",
		LastName:   "User",
		Bio:        "This is a test user bio",
		Attributes: map[string]any{"theme": "dark"},
	}
}

// ValidAuthParams returns valid authentication parameters.
func ValidAuthParams() *params.AuthParams {
	return &params.AuthParams{
		Identifier:        "testuser@example.com",
		IdentifierType:    "email",
		Password:          "SecurePassword123!",
		DeviceFingerprint: "test-device-fingerprint",
		DeviceName:        "Test Device",
		DeviceIP:          "127.0.0.1",
	}
}

// ValidPasswordChangeParam returns valid password change parameters.
func ValidPasswordChangeParam() *params.UserChangePasswordParam {
	return &params.UserChangePasswordParam{
		CurrentPassword: "SecurePassword123!",
		NewPassword:     "NewSecurePassword456!",
	}
}

// ValidUserProfileUpdateParam returns valid profile update parameters.
func ValidUserProfileUpdateParam() *params.UserProfileUpdateParam {
	firstName := "Updated"
	lastName := "Name"
	bio := "Updated bio"

	return &params.UserProfileUpdateParam{
		FirstName: &firstName,
		LastName:  &lastName,
		Bio:       &bio,
	}
}

// TestUserWithPassword combines a User model with plain-text password.
// This is useful for tests that need both the model and the password for authentication.
type TestUserWithPassword struct {
	User     *model.User
	Password string
}

// ValidTestUserWithPassword returns a TestUserWithPassword with a valid user and password.
func ValidTestUserWithPassword() *TestUserWithPassword {
	return &TestUserWithPassword{
		User:     ValidUserModel(),
		Password: "SecurePassword123!",
	}
}

// RandomTestUserWithPassword generates a user with random data and password.
func RandomTestUserWithPassword(suffix int) *TestUserWithPassword {
	user := &model.User{
		UID:      fmt.Sprintf("random-user-%d", suffix),
		Username: fmt.Sprintf("randomuser_%d", suffix),
		Email:    fmt.Sprintf("randomuser_%d@example.com", suffix),
		Password: "", // Should be hashed
		Status:   model.UserStatusActive,
	}

	return &TestUserWithPassword{
		User:     user,
		Password: "RandomPassword123!",
	}
}
