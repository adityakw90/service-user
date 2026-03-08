package grpc

import (
	"net"
	"sync"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/handler"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/middleware"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-user/internal/core/port/service"

	authpb "github.com/adityakw90/service-user-proto/gen/go/auth"
	devicepb "github.com/adityakw90/service-user-proto/gen/go/device"
	userpb "github.com/adityakw90/service-user-proto/gen/go/user"
	userFilepb "github.com/adityakw90/service-user-proto/gen/go/user_file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server          *grpc.Server
	listener        net.Listener
	authHandler     *handler.AuthHandler
	userHandler     *handler.UserHandler
	deviceHandler   *handler.DeviceHandler
	userFileHandler *handler.UserFileHandler
	m               *monitoring.Monitoring
	regOnce         sync.Once
}

func NewServer(
	authService service.AuthService,
	userService service.UserService,
	deviceService service.DeviceService,
	userFileService service.UserFileService,
	mon *monitoring.Monitoring,
) *Server {
	// Create validator instance
	v := validator.New()

	// Create handlers with validator injection
	authHandler := handler.NewAuthHandler(authService, v)
	userHandler := handler.NewUserHandler(userService, v)
	deviceHandler := handler.NewDeviceHandler(deviceService, v)
	userFileHandler := handler.NewUserFileHandler(userFileService)

	// Create gRPC server with chained interceptors
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			middleware.ChainUnaryInterceptors(
				middleware.UnaryRequestInterceptor(mon),
			),
		),
		grpc.StreamInterceptor(
			middleware.ChainStreamInterceptors(
				middleware.StreamRequestInterceptor(mon),
			),
		),
	)

	return &Server{
		server:          server,
		authHandler:     authHandler,
		userHandler:     userHandler,
		deviceHandler:   deviceHandler,
		userFileHandler: userFileHandler,
		m:               mon,
	}
}

func (s *Server) RegisterServices() {
	s.regOnce.Do(func() {
		authpb.RegisterAuthServiceServer(s.server, s.authHandler)
		userpb.RegisterUserServiceServer(s.server, s.userHandler)
		devicepb.RegisterDeviceServiceServer(s.server, s.deviceHandler)
		userFilepb.RegisterUserFileServiceServer(s.server, s.userFileHandler)
		reflection.Register(s.server)
	})
}

func (s *Server) Start(address string) error {
	var err error
	s.listener, err = net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.m.Logger.Info("gRPC server listening", map[string]interface{}{
		"addr": address,
	})

	s.RegisterServices()
	s.m.Logger.Info("register service", map[string]interface{}{
		"addr": address,
	})

	return s.server.Serve(s.listener)
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
	if s.listener != nil {
		s.listener.Close()
	}
}

// Addr returns the server address.
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}

// GetAuthHandler returns the auth handler.
func (s *Server) GetAuthHandler() *handler.AuthHandler {
	return s.authHandler
}

// GetUserHandler returns the user handler.
func (s *Server) GetUserHandler() *handler.UserHandler {
	return s.userHandler
}

// GetDeviceHandler returns the device handler.
func (s *Server) GetDeviceHandler() *handler.DeviceHandler {
	return s.deviceHandler
}

// GetUserFileHandler returns the user file handler.
func (s *Server) GetUserFileHandler() *handler.UserFileHandler {
	return s.userFileHandler
}
