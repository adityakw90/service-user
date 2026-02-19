package response

import (
	"context"
	"errors"
	"fmt"

	grpcValidator "github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	domainErrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/go-playground/validator/v10"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

func MakeErrorResponse(err error) error {
	// Handle CustomError type
	var customErr *domainErrors.CustomError
	if errors.As(err, &customErr) {
		return buildResponse(customErr)
	}

	// Handle request timeout or cancellation
	if errors.Is(err, context.Canceled) {
		return buildResponse(domainErrors.ErrRequestCanceled)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return buildResponse(domainErrors.ErrRequestTimeout)
	}

	// Handle gRPC errors (*status.Status)
	if _, ok := status.FromError(err); ok {
		return err
	}

	// Handle validation errors
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		newErr := ValidationError(validationErrs)
		return buildResponse(newErr)
	}

	// Default: Internal server error
	return buildResponse(domainErrors.ErrInternalServerError)
}

func buildResponse(err *domainErrors.CustomError) error {
	statusCode := extractErrorCode(err.Code)

	// Mask sensitive messages
	errMessage := err.Message
	switch err.Code {
	case domainErrors.ErrTraceInformationMissing.Code:
		errMessage = domainErrors.ErrInternalServerError.Message
	}

	st := status.New(statusCode, errMessage)

	// Build BadRequest field violations from ErrorMap
	var fieldViolations []*errdetails.BadRequest_FieldViolation
	for field, messages := range err.Errors {
		for _, msg := range messages {
			fieldViolations = append(fieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       field,
				Description: msg,
			})
		}
	}

	// Add structured error info
	badReq := &errdetails.BadRequest{
		FieldViolations: fieldViolations,
	}
	customInfo := &errdetails.ErrorInfo{
		Reason: "SERVICE_ERROR",
		Domain: "user.service", // customize if needed
		Metadata: map[string]string{
			"code":    fmt.Sprintf("%d", err.Code),
			"message": err.Message,
		},
	}

	stWithDetails, detailErr := st.WithDetails(badReq, customInfo)
	if detailErr != nil {
		// fallback to regular gRPC error if detail marshalling fails
		return status.Errorf(statusCode, "%s", errMessage)
	}

	return stWithDetails.Err()
}

func ValidationError(err validator.ValidationErrors) *domainErrors.CustomError {
	return domainErrors.NewCustomError(
		domainErrors.ErrValidation.Code,
		domainErrors.ErrValidation.Message,
		extractValidationErrors(err),
	)
}

// extractValidationErrors converts validator errors into a Marshmallow-style structure.
func extractValidationErrors(errs validator.ValidationErrors) domainErrors.ErrorMap {
	errMap := make(domainErrors.ErrorMap)

	for _, fieldErr := range errs {
		field := fieldErr.Field()
		message := grpcValidator.FormatFieldError(fieldErr)

		// Append error messages as a list per field
		errMap[field] = append(errMap[field], message)
	}

	return errMap
}
