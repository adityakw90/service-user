package middleware

import (
	"context"
	"errors"
	"testing"

	monitoring "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/response"
	domainErrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TestUnaryRequestInterceptor_ErrorConversion verifies that the middleware properly
// converts domain errors to gRPC errors using MakeErrorResponse
func TestUnaryRequestInterceptor_ErrorConversion(t *testing.T) {
	tests := []struct {
		name          string
		inputError    error
		expectCode    codes.Code
		expectMessage string
	}{
		{
			name:          "CustomError - NotFound",
			inputError:    domainErrors.ErrNotFound,
			expectCode:    codes.NotFound,
			expectMessage: "resource not found",
		},
		{
			name:          "CustomError - InvalidArgument",
			inputError:    domainErrors.ErrInvalidArgument,
			expectCode:    codes.InvalidArgument,
			expectMessage: "invalid argument",
		},
		{
			name:          "CustomError - PermissionDenied",
			inputError:    domainErrors.ErrPermissionDenied,
			expectCode:    codes.PermissionDenied,
			expectMessage: "permission denied",
		},
		{
			name:          "CustomError - InternalServerError",
			inputError:    domainErrors.ErrInternalServerError,
			expectCode:    codes.Internal,
			expectMessage: "internal server error",
		},
		{
			name:          "CustomError - Validation",
			inputError:    domainErrors.ErrValidation,
			expectCode:    codes.InvalidArgument,
			expectMessage: "validation error",
		},
		{
			name:          "Standard error - Default conversion",
			inputError:    errors.New("unexpected error"),
			expectCode:    codes.Internal,
			expectMessage: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock monitoring
			mockMonitoring := setupMockMonitoring(t)

			// Create interceptor
			interceptor := UnaryRequestInterceptor(mockMonitoring)

			// Create handler that returns the test error
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, tt.inputError
			}

			// Execute request
			_, err := interceptor(
				context.Background(),
				"test-request",
				&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"},
				handler,
			)

			// Verify error was returned
			require.Error(t, err, "Expected error to be returned")

			// Verify error is a gRPC status error
			st, ok := status.FromError(err)
			require.True(t, ok, "Expected gRPC status error")

			// Verify status code matches expected
			assert.Equal(t, tt.expectCode, st.Code(), "Expected status code %v, got %v", tt.expectCode, st.Code())

			// Verify message contains expected text
			assert.Contains(t, st.Message(), tt.expectMessage, "Expected message to contain '%s'", tt.expectMessage)
		})
	}
}

// TestUnaryRequestInterceptor_MakeErrorResponseIntegration tests that the middleware
// properly integrates with MakeErrorResponse from the response package
func TestUnaryRequestInterceptor_MakeErrorResponseIntegration(t *testing.T) {
	mockMonitoring := setupMockMonitoring(t)
	interceptor := UnaryRequestInterceptor(mockMonitoring)

	// Test that MakeErrorResponse is called correctly
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Return a domain error that should be converted
		return nil, domainErrors.ErrNotFound
	}

	_, err := interceptor(
		context.Background(),
		"test-request",
		&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"},
		handler,
	)

	require.Error(t, err)

	// Verify that the error returned matches what MakeErrorResponse would produce
	expectedErr := response.MakeErrorResponse(domainErrors.ErrNotFound)
	expectedStatus, _ := status.FromError(expectedErr)
	actualStatus, _ := status.FromError(err)

	assert.Equal(t, expectedStatus.Code(), actualStatus.Code())
	assert.Equal(t, expectedStatus.Message(), actualStatus.Message())
}

type testServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *testServerStream) Context() context.Context { return s.ctx }

func TestRequestInterceptors_ActorNamePropagation(t *testing.T) {
	md := metadata.Pairs("client", "web-app", "actor-id", "user-123", "actor-type", "user", "actor-name", "Alice")
	incomingCtx := metadata.NewIncomingContext(context.Background(), md)

	t.Run("unary", func(t *testing.T) {
		interceptor := UnaryRequestInterceptor(setupMockMonitoring(t))
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			actorID, actorType, actorName := util.GetActor(ctx)
			assert.Equal(t, "user-123", actorID)
			assert.Equal(t, "user", actorType)
			assert.Equal(t, "Alice", actorName)
			return nil, nil
		}
		_, err := interceptor(incomingCtx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Service/Unary"}, handler)
		require.NoError(t, err)
	})

	t.Run("stream", func(t *testing.T) {
		interceptor := StreamRequestInterceptor(setupMockMonitoring(t))
		stream := &testServerStream{ctx: incomingCtx}
		handler := func(srv interface{}, ss grpc.ServerStream) error {
			actorID, actorType, actorName := util.GetActor(ss.Context())
			assert.Equal(t, "user-123", actorID)
			assert.Equal(t, "user", actorType)
			assert.Equal(t, "Alice", actorName)
			return nil
		}
		err := interceptor(nil, stream, &grpc.StreamServerInfo{FullMethod: "/test.Service/Stream"}, handler)
		require.NoError(t, err)
	})
}

// setupMockMonitoring creates a minimal mock monitoring setup for testing
func setupMockMonitoring(t *testing.T) *monitoring.Monitoring {
	t.Helper()

	// Create a real monitoring instance with stdout providers for simpler testing
	m, err := infra.NewMonitoring(&infra.MonitoringConfig{
		ServiceName:         "test-service",
		Environment:         "test",
		InstanceName:        "test-instance",
		InstanceHost:        "localhost",
		LoggerLevel:         "error", // Only log errors to reduce noise
		LoggerCallerSkipNum: 1,
		TracerProvider:      "stdout", // Use stdout for simplicity
		TracerProviderHost:  "localhost",
		TracerProviderPort:  6831,
		TracerSampleRatio:   1.0,
		MetricProvider:      "stdout", // Use stdout for simplicity
	})
	if err != nil {
		t.Fatalf("failed to create monitoring: %v", err)
	}

	// Cleanup monitoring when test completes
	t.Cleanup(func() {
		m.Shutdown(context.Background())
	})

	return m
}
