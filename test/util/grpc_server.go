package testutil

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	authgrpc "github.com/adityakw90/service-user-proto/gen/go/auth"
	devicegrpc "github.com/adityakw90/service-user-proto/gen/go/device"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	user_filegrpc "github.com/adityakw90/service-user-proto/gen/go/user_file"

	grpcadapter "github.com/adityakw90/service-user/internal/adapter/api/grpc/handler"
	grpcvalidator "github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
)

// TestGRPCServer wraps a running gRPC server for testing.
type TestGRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	addr     string

	// gRPC handlers (for direct access if needed)
	userHandler     *grpcadapter.UserHandler
	authHandler     *grpcadapter.AuthHandler
	deviceHandler   *grpcadapter.DeviceHandler
	userFileHandler *grpcadapter.UserFileHandler
}

// NewTestGRPCServer creates and starts a test gRPC server.
// Uses the provided TestServices for all dependencies.
func NewTestGRPCServer(testServices *TestServices) (*TestGRPCServer, error) {
	// Start listener on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC handlers
	validator := grpcvalidator.New()
	userHandler := grpcadapter.NewUserHandler(testServices.UserService, validator)
	authHandler := grpcadapter.NewAuthHandler(testServices.AuthService, validator)
	deviceHandler := grpcadapter.NewDeviceHandler(testServices.DeviceService, validator)
	userFileHandler := grpcadapter.NewUserFileHandler(testServices.UserFileService, validator)

	// Create server
	server := grpc.NewServer()
	usergrpc.RegisterUserServiceServer(server, userHandler)
	authgrpc.RegisterAuthServiceServer(server, authHandler)
	devicegrpc.RegisterDeviceServiceServer(server, deviceHandler)
	user_filegrpc.RegisterUserFileServiceServer(server, userFileHandler)
	reflection.Register(server)

	// Start server in background
	go func() {
		if err := server.Serve(listener); err != nil {
			// Server stopped
		}
	}()

	return &TestGRPCServer{
		server:          server,
		listener:        listener,
		addr:            listener.Addr().String(),
		userHandler:     userHandler,
		authHandler:     authHandler,
		deviceHandler:   deviceHandler,
		userFileHandler: userFileHandler,
	}, nil
}

// Addr returns the server address.
func (s *TestGRPCServer) Addr() string {
	return s.addr
}

// Close stops the server.
func (s *TestGRPCServer) Close() {
	if s.server != nil {
		s.server.GracefulStop()
	}
	if s.listener != nil {
		s.listener.Close()
	}
}
