package service

import (
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/pkg/util"
)

// Helper function to create a test user
func createTestUser(id int64, uid, username, email, password string, status model.UserStatus) *model.User {
	now := time.Now().UTC()
	return &model.User{
		ID:        id,
		UID:       uid,
		Username:  username,
		Email:     email,
		Password:  password,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Helper function to create a test user profile
func createTestProfile(userID int64, userUID string) *model.UserProfile {
	now := time.Now().UTC()
	return &model.UserProfile{
		UserID:    userID,
		UserUID:   userUID,
		FirstName: "Test",
		LastName:  "User",
		Bio:       "Test bio",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Helper function to create a test device
func createTestDevice(id int64, uid, name, fingerprint string) *model.Device {
	now := time.Now().UTC()
	return &model.Device{
		ID:                id,
		UID:               uid,
		DeviceName:        name,
		DeviceFingerprint: fingerprint,
		CreatedAt:         now,
	}
}

// Helper function to create a test user device
func createUserDevice(userID, deviceID int64, ipAddress string) *model.UserDevice {
	now := time.Now().UTC()
	return &model.UserDevice{
		UserID:    userID,
		DeviceID:  deviceID,
		IPAddress: ipAddress,
		CreatedAt: now,
	}
}

// Helper function to create a test user file
func createUserFile(id int64, uid, userUID, fileType string) *model.UserFile {
	now := time.Now().UTC()
	return &model.UserFile{
		ID:         id,
		UID:        uid,
		UserID:     1,
		UserUID:    userUID,
		FileType:   fileType,
		FileName:   "test.jpg",
		FilePath:   "/uploads/test.jpg",
		MimeType:   "image/jpeg",
		SizeBytes:  1024,
		Visibility: model.FileVisibilityPrivate,
		CreatedAt:  now,
	}
}

// Helper function to create a test user pin
func createUserPin(userID int64, userUID string, code string) *model.UserPin {
	now := time.Now().UTC()
	return &model.UserPin{
		UserID:    userID,
		UserUID:   userUID,
		Code:      code,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Helper function to create auth params
func createAuthParams(identifier, identifierType, password, deviceName, deviceFingerprint, deviceIP string) *param.AuthParams {
	return &param.AuthParams{
		Identifier:        identifier,
		IdentifierType:    identifierType,
		Password:          password,
		DeviceName:        util.Ptr(deviceName),
		DeviceFingerprint: util.Ptr(deviceFingerprint),
		DeviceIP:          util.Ptr(deviceIP),
	}
}

// Helper function to create user create params
func createUserCreateParams(username, email, password string) *param.UserCreateParam {
	return &param.UserCreateParam{
		Username: username,
		Email:    email,
		Password: password,
	}
}

// Helper function to create user update params
func createUserUpdateParams(username, email, password *string, status *model.UserStatus) *param.UserUpdateParam {
	return &param.UserUpdateParam{
		Username: username,
		Email:    email,
		Password: password,
		Status:   status,
	}
}

// Helper function to create a deleted user
func createDeletedUser(id int64, uid, username, email string) *model.User {
	now := time.Now().UTC()
	return &model.User{
		ID:        id,
		UID:       uid,
		Username:  username,
		Email:     email,
		Password:  "hashed_password",
		Status:    model.UserStatusActive,
		DeletedAt: &now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
