package response

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
)

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
