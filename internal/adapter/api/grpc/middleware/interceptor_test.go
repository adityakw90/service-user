package middleware

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockServerStream is a mock implementation of grpc.ServerStream for testing.
type mockServerStream struct {
	grpc.ServerStream
}

// TestChainUnaryInterceptors tests the ChainUnaryInterceptors function.
func TestChainUnaryInterceptors(t *testing.T) {
	tests := []struct {
		name         string
		interceptors []grpc.UnaryServerInterceptor
		wantCalls    []string
		wantResponse interface{}
		wantErr      bool
	}{
		{
			name: "Happy Path - two interceptors chain correctly",
			interceptors: nil, // Will be created in test
			wantCalls:    []string{"interceptor1", "interceptor2", "handler"},
			wantResponse: "response",
			wantErr:      false,
		},
		{
			name: "Single Interceptor - directly returns the interceptor",
			interceptors: nil,
			wantCalls:    []string{"interceptor1", "handler"},
			wantResponse: "response",
			wantErr:      false,
		},
		{
			name:         "No Interceptors - directly calls handler",
			interceptors: []grpc.UnaryServerInterceptor{},
			wantCalls:    []string{"handler"},
			wantResponse: "response",
			wantErr:      false,
		},
		{
			name:         "Multiple Interceptors - three interceptors chain correctly",
			interceptors: nil,
			wantCalls:    []string{"interceptor1", "interceptor2", "interceptor3", "handler"},
			wantResponse: "response",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := []string{}

			// Create a tracking handler that records when it's called
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				calls = append(calls, "handler")
				return tt.wantResponse, nil
			}

			// Create interceptors inline to avoid closure issues
			var interceptors []grpc.UnaryServerInterceptor
			if tt.interceptors != nil {
				interceptors = tt.interceptors
			} else {
				// Create tracking interceptors based on test case
				switch tt.name {
				case "Happy Path - two interceptors chain correctly":
					interceptors = []grpc.UnaryServerInterceptor{
						func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
							calls = append(calls, "interceptor1")
							return handler(ctx, req)
						},
						func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
							calls = append(calls, "interceptor2")
							return handler(ctx, req)
						},
					}
				case "Single Interceptor - directly returns the interceptor":
					interceptors = []grpc.UnaryServerInterceptor{
						func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
							calls = append(calls, "interceptor1")
							return handler(ctx, req)
						},
					}
				case "Multiple Interceptors - three interceptors chain correctly":
					interceptors = []grpc.UnaryServerInterceptor{
						func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
							calls = append(calls, "interceptor1")
							return handler(ctx, req)
						},
						func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
							calls = append(calls, "interceptor2")
							return handler(ctx, req)
						},
						func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
							calls = append(calls, "interceptor3")
							return handler(ctx, req)
						},
					}
				}
			}

			chained := ChainUnaryInterceptors(interceptors...)
			result, err := chained(context.Background(), "request", &grpc.UnaryServerInfo{}, handler)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResponse, result)
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

// TestChainStreamInterceptors tests the ChainStreamInterceptors function.
func TestChainStreamInterceptors(t *testing.T) {
	tests := []struct {
		name         string
		interceptors []grpc.StreamServerInterceptor
		wantCalls    []string
		wantErr      bool
	}{
		{
			name:         "Happy Path - two interceptors chain correctly",
			interceptors: nil,
			wantCalls:    []string{"interceptor1", "interceptor2", "handler"},
			wantErr:      false,
		},
		{
			name:         "Single Interceptor - directly returns the interceptor",
			interceptors: nil,
			wantCalls:    []string{"interceptor1", "handler"},
			wantErr:      false,
		},
		{
			name:         "No Interceptors - directly calls handler",
			interceptors: []grpc.StreamServerInterceptor{},
			wantCalls:    []string{"handler"},
			wantErr:      false,
		},
		{
			name:         "Multiple Interceptors - three interceptors chain correctly",
			interceptors: nil,
			wantCalls:    []string{"interceptor1", "interceptor2", "interceptor3", "handler"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := []string{}

			// Create a tracking handler that records when it's called
			handler := func(srv interface{}, ss grpc.ServerStream) error {
				calls = append(calls, "handler")
				return nil
			}

			// Create interceptors inline to avoid closure issues
			var interceptors []grpc.StreamServerInterceptor
			if tt.interceptors != nil {
				interceptors = tt.interceptors
			} else {
				// Create tracking interceptors based on test case
				switch tt.name {
				case "Happy Path - two interceptors chain correctly":
					interceptors = []grpc.StreamServerInterceptor{
						func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
							calls = append(calls, "interceptor1")
							return handler(srv, ss)
						},
						func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
							calls = append(calls, "interceptor2")
							return handler(srv, ss)
						},
					}
				case "Single Interceptor - directly returns the interceptor":
					interceptors = []grpc.StreamServerInterceptor{
						func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
							calls = append(calls, "interceptor1")
							return handler(srv, ss)
						},
					}
				case "Multiple Interceptors - three interceptors chain correctly":
					interceptors = []grpc.StreamServerInterceptor{
						func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
							calls = append(calls, "interceptor1")
							return handler(srv, ss)
						},
						func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
							calls = append(calls, "interceptor2")
							return handler(srv, ss)
						},
						func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
							calls = append(calls, "interceptor3")
							return handler(srv, ss)
						},
					}
				}
			}

			chained := ChainStreamInterceptors(interceptors...)
			err := chained("service", &mockServerStream{}, &grpc.StreamServerInfo{}, handler)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}
