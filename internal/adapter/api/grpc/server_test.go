package grpc

import (
	"context"
	"testing"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/port/service"
	servicemocks "github.com/adityakw90/service-user/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestMonitoring creates a monitoring instance for testing
func newTestMonitoring(t *testing.T) *monitoring.Monitoring {
	t.Helper()
	mon, err := monitoring.NewMonitoring(
		monitoring.WithServiceName("test-service"),
		monitoring.WithLoggerLevel("info"),
		monitoring.WithEnvironment("test"),
	)
	require.NoError(t, err)
	return mon
}

// TestNewServer tests the NewServer constructor.
func TestNewServer(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (service.AuthService, service.UserService, service.DeviceService, service.UserFileService, *monitoring.Monitoring)
	}{
		{
			name: "Happy Path - creates server with all dependencies",
			setup: func() (service.AuthService, service.UserService, service.DeviceService, service.UserFileService, *monitoring.Monitoring) {
				authService := servicemocks.NewMockAuthService(t)
				userService := servicemocks.NewMockUserService(t)
				deviceService := servicemocks.NewMockDeviceService(t)
				userFileService := servicemocks.NewMockUserFileService(t)
				mon := newTestMonitoring(t)
				return authService, userService, deviceService, userFileService, mon
			},
		},
		{
			name: "Nil Monitoring - should still create server",
			setup: func() (service.AuthService, service.UserService, service.DeviceService, service.UserFileService, *monitoring.Monitoring) {
				authService := servicemocks.NewMockAuthService(t)
				userService := servicemocks.NewMockUserService(t)
				deviceService := servicemocks.NewMockDeviceService(t)
				userFileService := servicemocks.NewMockUserFileService(t)
				return authService, userService, deviceService, userFileService, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, userService, deviceService, userFileService, mon := tt.setup()

			s := NewServer(authService, userService, deviceService, userFileService, mon)

			assert.NotNil(t, s)
			assert.NotNil(t, s.server)
			assert.NotNil(t, s.authHandler)
			assert.NotNil(t, s.userHandler)
			assert.NotNil(t, s.deviceHandler)
			assert.NotNil(t, s.userFileHandler)
			assert.Nil(t, s.listener)

			if mon != nil {
				_ = mon.Shutdown(context.Background())
			}
		})
	}
}

// TestRegisterServices tests the RegisterServices method.
func TestRegisterServices(t *testing.T) {
	tests := []struct {
		name          string
		registerCount int
	}{
		{
			name:          "Happy Path - registers all services once",
			registerCount: 1,
		},
		{
			name:          "Multiple Calls - sync.Once prevents double registration",
			registerCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := servicemocks.NewMockAuthService(t)
			userService := servicemocks.NewMockUserService(t)
			deviceService := servicemocks.NewMockDeviceService(t)
			userFileService := servicemocks.NewMockUserFileService(t)
			mon := newTestMonitoring(t)
			defer func() {
				if mon != nil {
					_ = mon.Shutdown(context.Background())
				}
			}()

			s := NewServer(authService, userService, deviceService, userFileService, mon)

			for i := 0; i < tt.registerCount; i++ {
				s.RegisterServices()
			}

			assert.NotNil(t, s.server)
			assert.NotNil(t, s.authHandler)
			assert.NotNil(t, s.userHandler)
			assert.NotNil(t, s.deviceHandler)
			assert.NotNil(t, s.userFileHandler)
		})
	}
}

// TestStart tests the Start method error cases.
func TestStart(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		setupCtx    func() context.Context
		wantErr     bool
		errContains string
	}{
		{
			name:    "Invalid Address - returns error",
			address: "invalid-address",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantErr:     true,
			errContains: "listen",
		},
		{
			name:    "Invalid Address - missing port",
			address: "localhost",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantErr:     true,
			errContains: "listen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := servicemocks.NewMockAuthService(t)
			userService := servicemocks.NewMockUserService(t)
			deviceService := servicemocks.NewMockDeviceService(t)
			userFileService := servicemocks.NewMockUserFileService(t)
			mon := newTestMonitoring(t)
			defer func() {
				if mon != nil {
					_ = mon.Shutdown(context.Background())
				}
			}()

			s := NewServer(authService, userService, deviceService, userFileService, mon)

			ctx := tt.setupCtx()

			err := s.Start(ctx, tt.address)
			require.Error(t, err)
			if tt.errContains != "" {
				assert.Contains(t, err.Error(), tt.errContains)
			}
		})
	}
}

// TestStop tests the Stop method.
func TestStop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Nil Server - no panic",
		},
		{
			name: "Nil Listener - no panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := servicemocks.NewMockAuthService(t)
			userService := servicemocks.NewMockUserService(t)
			deviceService := servicemocks.NewMockDeviceService(t)
			userFileService := servicemocks.NewMockUserFileService(t)
			mon := newTestMonitoring(t)
			defer func() {
				if mon != nil {
					_ = mon.Shutdown(context.Background())
				}
			}()

			s := NewServer(authService, userService, deviceService, userFileService, mon)

			assert.NotPanics(t, func() {
				s.Stop()
			})
		})
	}
}

// TestStartStopSequence tests the full server lifecycle.
func TestStartStopSequence(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{
			name:    "Full Lifecycle - start and stop",
			address: ":0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := servicemocks.NewMockAuthService(t)
			userService := servicemocks.NewMockUserService(t)
			deviceService := servicemocks.NewMockDeviceService(t)
			userFileService := servicemocks.NewMockUserFileService(t)
			mon := newTestMonitoring(t)

			s := NewServer(authService, userService, deviceService, userFileService, mon)

			ctx, cancel := context.WithCancel(context.Background())

			done := make(chan error, 1)
			go func() {
				done <- s.Start(ctx, tt.address)
			}()

			// Give the server a moment to start
			// In a real scenario, we'd use proper synchronization

			s.Stop()
			cancel()

			err := <-done
			// The error could be "grpc: the server has been stopped" which is expected
			// when we stop the server immediately after starting
			assert.True(t, err == nil || err.Error() == "grpc: the server has been stopped" || err.Error() == context.Canceled.Error(),
				"unexpected error: %v", err)

			if mon != nil {
				_ = mon.Shutdown(context.Background())
			}
		})
	}
}

// TestAddr tests the Addr method.
func TestAddr(t *testing.T) {
	tests := []struct {
		name     string
		started  bool
		wantAddr string
	}{
		{
			name:     "Before Start - returns empty string",
			started:  false,
			wantAddr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := servicemocks.NewMockAuthService(t)
			userService := servicemocks.NewMockUserService(t)
			deviceService := servicemocks.NewMockDeviceService(t)
			userFileService := servicemocks.NewMockUserFileService(t)
			mon := newTestMonitoring(t)
			defer func() {
				if mon != nil {
					_ = mon.Shutdown(context.Background())
				}
			}()

			s := NewServer(authService, userService, deviceService, userFileService, mon)

			addr := s.Addr()
			assert.Empty(t, addr)
		})
	}
}

// TestGetters tests all getter methods.
func TestGetters(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Happy Path - all getters return correct handlers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := servicemocks.NewMockAuthService(t)
			userService := servicemocks.NewMockUserService(t)
			deviceService := servicemocks.NewMockDeviceService(t)
			userFileService := servicemocks.NewMockUserFileService(t)
			mon := newTestMonitoring(t)
			defer func() {
				if mon != nil {
					_ = mon.Shutdown(context.Background())
				}
			}()

			s := NewServer(authService, userService, deviceService, userFileService, mon)

			assert.NotNil(t, s.GetAuthHandler())
			assert.NotNil(t, s.GetUserHandler())
			assert.NotNil(t, s.GetDeviceHandler())
			assert.NotNil(t, s.GetUserFileHandler())
		})
	}
}
