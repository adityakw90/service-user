package response

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	device "github.com/adityakw90/service-user-proto/gen/go/device"
	user "github.com/adityakw90/service-user-proto/gen/go/user"
	userFile "github.com/adityakw90/service-user-proto/gen/go/user_file"
	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

func ToProtoUser(u *model.User) *user.User {
	if u == nil {
		return nil
	}
	return &user.User{
		Uid:       u.UID,
		Username:  u.Username,
		Email:     u.Email,
		Status:    int32(u.Status),
		CreatedAt: Timestamp(u.CreatedAt),
		UpdatedAt: Timestamp(u.UpdatedAt),
		DeletedAt: TimestampPtr(u.DeletedAt),
	}
}

func ToProtoProfile(p *model.UserProfile) *user.Profile {
	if p == nil {
		return nil
	}
	return &user.Profile{
		Uid:        "",
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		Bio:        p.Bio,
		Attributes: ToStruct(p.Attributes),
	}
}

func ToProtoDevice(d *model.Device) *user.Device {
	if d == nil {
		return nil
	}
	return &user.Device{
		DeviceUid:  d.UID,
		DeviceName: d.DeviceName,
		CreatedAt:  Timestamp(d.CreatedAt),
	}
}

func ToProtoDeviceFull(d *model.Device) *device.Device {
	if d == nil {
		return nil
	}
	return &device.Device{
		Uid:               d.UID,
		DeviceFingerprint: d.DeviceFingerprint,
		DeviceName:        d.DeviceName,
		CreatedAt:         Timestamp(d.CreatedAt),
	}
}

func ToProtoUserFile(f *model.UserFile) *userFile.UserFile {
	if f == nil {
		return nil
	}
	return &userFile.UserFile{
		Uid:       f.UID,
		UserUid:   f.UserUID,
		FileType:  f.FileType,
		FileName:  f.FileName,
		FilePath:  f.FilePath,
		MimeType:  f.MimeType,
		SizeBytes: f.SizeBytes,
		Visibility: f.Visibility,
		CreatedAt: Timestamp(f.CreatedAt),
	}
}

func Timestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func TimestampPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// MapAuthError maps service errors to gRPC errors for AuthHandler.
func MapAuthError(err error) error {
	if err == nil {
		return nil
	}
	switch err {
	case domainerrors.ErrInvalidCredentials:
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case domainerrors.ErrTokenInvalid:
		return status.Error(codes.Unauthenticated, "invalid token")
	case domainerrors.ErrTokenExpired:
		return status.Error(codes.Unauthenticated, "token has expired")
	case domainerrors.ErrDeviceNotFound:
		return status.Error(codes.NotFound, "device not found")
	case domainerrors.ErrUserNotFound:
		return status.Error(codes.NotFound, "user not found")
	case domainerrors.ErrUserDeleted:
		return status.Error(codes.FailedPrecondition, "user has been deleted")
	case domainerrors.ErrPinNotSet:
		return status.Error(codes.FailedPrecondition, "PIN not set for user")
	default:
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}
}

// MapError maps general user service errors to gRPC errors.
func MapError(err error) error {
	switch err {
	case domainerrors.ErrUserNotFound:
		return status.Error(codes.NotFound, "user not found")
	case domainerrors.ErrDuplicateEmail:
		return status.Error(codes.AlreadyExists, "email already exists")
	case domainerrors.ErrDuplicateUsername:
		return status.Error(codes.AlreadyExists, "username already exists")
	case domainerrors.ErrUserDeleted:
		return status.Error(codes.FailedPrecondition, "user has been deleted")
	case domainerrors.ErrProfileNotFound:
		return status.Error(codes.NotFound, "profile not found")
	case domainerrors.ErrDeviceNotFound:
		return status.Error(codes.NotFound, "device not found")
	case domainerrors.ErrDeviceRevoked:
		return status.Error(codes.PermissionDenied, "device access has been revoked")
	case domainerrors.ErrPinNotSet:
		return status.Error(codes.FailedPrecondition, "PIN not set for user")
	default:
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}
}
