package response

import (
	domainErrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"google.golang.org/grpc/codes"
)

var GrpcErrorCode = map[int]codes.Code{
	domainErrors.ErrInternalServerError.Code:     codes.Internal,
	domainErrors.ErrTraceInformationMissing.Code: codes.Internal,
	domainErrors.ErrRequestCanceled.Code:         codes.Canceled,
	domainErrors.ErrRequestTimeout.Code:          codes.DeadlineExceeded,
	domainErrors.ErrRequestAborted.Code:          codes.Aborted,
	domainErrors.ErrUnimplemented.Code:           codes.Unimplemented,
	domainErrors.ErrNotFound.Code:                codes.NotFound,
	domainErrors.ErrInvalidArgument.Code:         codes.InvalidArgument,
	domainErrors.ErrValidation.Code:              codes.InvalidArgument,
	domainErrors.ErrPermissionDenied.Code:        codes.PermissionDenied,
	domainErrors.ErrResourceConflict.Code:        codes.Aborted,

	domainErrors.ErrInvalidCredentials.Code:    codes.InvalidArgument,
	domainErrors.ErrInvalidIdentifierType.Code: codes.InvalidArgument,

	// rate limiting errors
	domainErrors.ErrRateLimitExceeded.Code: codes.ResourceExhausted,

	// token errors
	domainErrors.ErrTokenRevoked.Code:      codes.InvalidArgument,
	domainErrors.ErrTokenBlacklisted.Code:  codes.InvalidArgument,
	domainErrors.ErrTokenInvalid.Code:      codes.InvalidArgument,
	domainErrors.ErrTokenInvalidClaim.Code: codes.InvalidArgument,
	domainErrors.ErrTokenExpired.Code:      codes.InvalidArgument,
	domainErrors.ErrInvalidTokenType.Code:  codes.InvalidArgument,

	// oauth errors
	domainErrors.ErrOAuthClientIDRequired.Code:           codes.InvalidArgument,
	domainErrors.ErrOAuthClientSecretRequired.Code:       codes.InvalidArgument,
	domainErrors.ErrOAuthFailedGenerateCodeVerifier.Code: codes.InvalidArgument,
	domainErrors.ErrOAuthInvalidMinVerifierLength.Code:   codes.InvalidArgument,
	domainErrors.ErrOAuthInvalidMaxVerifierLength.Code:   codes.InvalidArgument,
	domainErrors.ErrOAuthInvalidState.Code:               codes.InvalidArgument,
	domainErrors.ErrOAuthExchangeFailed.Code:             codes.InvalidArgument,
	domainErrors.ErrOAuthUserInfoFailed.Code:             codes.InvalidArgument,
	domainErrors.ErrOAuthAccessDenied.Code:               codes.InvalidArgument,
	domainErrors.ErrOAuthInvalidCode.Code:                codes.InvalidArgument,
	domainErrors.ErrOAuthServerError.Code:                codes.InvalidArgument,
	domainErrors.ErrOAuthTemporarilyUnavailable.Code:     codes.InvalidArgument,
	// PKCE-specific errors
	domainErrors.ErrOAuthCodeVerifierMissing.Code: codes.InvalidArgument,

	// account errors
	domainErrors.ErrAccountLockedOut.Code: codes.InvalidArgument,

	// user errors
	domainErrors.ErrInvalidUID.Code:             codes.InvalidArgument,
	domainErrors.ErrInvalidUsername.Code:        codes.InvalidArgument,
	domainErrors.ErrInvalidEmail.Code:           codes.InvalidArgument,
	domainErrors.ErrInvalidPassword.Code:        codes.InvalidArgument,
	domainErrors.ErrUserNotFound.Code:           codes.InvalidArgument,
	domainErrors.ErrUserAlreadyExists.Code:      codes.InvalidArgument,
	domainErrors.ErrDuplicateEmail.Code:         codes.InvalidArgument,
	domainErrors.ErrDuplicateUsername.Code:      codes.InvalidArgument,
	domainErrors.ErrInvalidStatus.Code:          codes.InvalidArgument,
	domainErrors.ErrUserDeleted.Code:            codes.InvalidArgument,
	domainErrors.ErrUserInactive.Code:           codes.InvalidArgument,
	domainErrors.ErrInvalidCurrentPassword.Code: codes.InvalidArgument,
	domainErrors.ErrPasswordMismatch.Code:       codes.InvalidArgument,

	// profile errors
	domainErrors.ErrProfileNotFound.Code: codes.InvalidArgument,

	// pin errors
	domainErrors.ErrPinNotSet.Code:          codes.InvalidArgument,
	domainErrors.ErrPinInvalid.Code:         codes.InvalidArgument,
	domainErrors.ErrPINTooManyAttempts.Code: codes.InvalidArgument,

	// file errors
	domainErrors.ErrFileNotFound.Code: codes.InvalidArgument,

	// device errors
	domainErrors.ErrDeviceNotFound.Code:     codes.InvalidArgument,
	domainErrors.ErrDeviceRevoked.Code:      codes.InvalidArgument,
	domainErrors.ErrUserDeviceNotFound.Code: codes.InvalidArgument,
}

// Extract the status code (first 3 digits)
func extractErrorCode(code int) codes.Code {
	if grpcCode, ok := GrpcErrorCode[code]; ok {
		return grpcCode
	}
	return codes.Internal
}
