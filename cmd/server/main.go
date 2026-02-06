package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adityakw90/service-user-proto/gen/go/auth"
	"github.com/adityakw90/service-user-proto/gen/go/device"
	"github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user-proto/gen/go/user_file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcadapter "github.com/adityakw90/service-user/internal/adapter/api/grpc/handler"
	"github.com/adityakw90/service-user/internal/adapter/monitoring"
	"github.com/adityakw90/service-user/internal/adapter/publisher"
	"github.com/adityakw90/service-user/internal/adapter/repository"
	"github.com/adityakw90/service-user/internal/adapter/security"
	"github.com/adityakw90/service-user/internal/config"
	"github.com/adityakw90/service-user/internal/core/port"
	"github.com/adityakw90/service-user/internal/core/service"
	"github.com/adityakw90/service-user/internal/infra"
)

func main() {
	// Initialize logger
	logger := monitoring.NewLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// start context
	ctx := context.Background()

	// Connect to PostgreSQL
	dbPool, err := infra.NewPostgreConnection(ctx, &infra.PostgreConfig{
		Host:                  cfg.Database.Host,
		Port:                  cfg.Database.Port,
		User:                  cfg.Database.User,
		Password:              cfg.Database.Password,
		Name:                  cfg.Database.Name,
		SslMode:               cfg.Database.SslMode,
		Timezone:              cfg.Database.Timezone,
		MinConns:              cfg.Database.MinConns,
		MinIdleConns:          cfg.Database.MinIdleConns,
		MaxConns:              cfg.Database.MaxConns,
		MaxConnIdleTime:       cfg.Database.MaxConnIdleTime,
		MaxConnLifetime:       cfg.Database.MaxConnLifetime,
		MaxConnLifetimeJitter: cfg.Database.MaxConnLifetimeJitter,
		HealthCheckPeriod:     cfg.Database.HealthCheckPeriod,
	})
	if err != nil {
		logger.Fatal("failed to connect to database", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer dbPool.Close()
	logger.Info("connected to database", nil)

	// Connect to Redis using infra layer
	redisClient, err := infra.NewRedisConnection(context.Background(), &infra.RedisConfig{
		Host:              cfg.Redis.Host,
		Port:              cfg.Redis.Port,
		User:              cfg.Redis.User,
		Password:          cfg.Redis.Password,
		DB:                cfg.Redis.DB,
		PoolSize:          cfg.Redis.PoolSize,
		PoolTimeout:       cfg.Redis.PoolTimeout,
		ConnectionIdleMin: cfg.Redis.ConnectionIdleMin,
	})
	if err != nil {
		logger.Fatal("failed to connect to redis", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer redisClient.Close()
	logger.Info("connected to redis", nil)

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool)
	profileRepo := repository.NewProfileRepository(dbPool)
	pinRepo := repository.NewPinRepository(dbPool)
	deviceRepo := repository.NewDeviceRepository(dbPool)
	userDeviceRepo := repository.NewUserDeviceRepository(dbPool)
	_ = repository.NewUserFileRepository(dbPool)

	// Initialize hashers
	passwordHasher, err := security.NewHasher(cfg.PasswordHasher.Type, map[string]any{
		"cost":        cfg.PasswordHasher.Cost,
		"salt":        cfg.PasswordHasher.Salt,
		"memory":      cfg.PasswordHasher.Memory,
		"iterations":  cfg.PasswordHasher.Iterations,
		"parallelism": cfg.PasswordHasher.Parallelism,
		"saltLength":  cfg.PasswordHasher.SaltLength,
		"keyLength":   cfg.PasswordHasher.KeyLength,
	})
	if err != nil {
		logger.Fatal("failed to initialize password hasher", map[string]interface{}{
			"error": err.Error(),
		})
	}

	pinHasher, err := security.NewHasher(cfg.PINHasher.Type, map[string]any{
		"cost":        14,
		"salt":        cfg.PINHasher.Salt,
		"memory":      cfg.PINHasher.Memory,
		"iterations":  4,
		"parallelism": cfg.PINHasher.Parallelism,
		"saltLength":  cfg.PINHasher.SaltLength,
		"keyLength":   cfg.PINHasher.KeyLength,
	})
	if err != nil {
		logger.Fatal("failed to initialize pin hasher", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize token generator
	tokenGen := security.NewJWTGenerator(
		cfg.Jwt.SecretKey,
		cfg.Jwt.AccessExpiry,
		cfg.Jwt.RefreshExpiry,
	)

	// Initialize cache adapters (using *redis.Client directly from infra)
	tokenBlacklist := security.NewTokenBlacklistAdapter(redisClient, "token-blacklist:", 24*time.Hour)
	tokenWhitelist := security.NewTokenWhitelistAdapter(redisClient, "token-whitelist:", 15*time.Minute)

	// Initialize event publisher
	var eventPublisher port.EventPublisher
	if cfg.EventPublisherEndpoint != "" {
		eventPublisher = publisher.NewHTTPPublisher(publisher.HttpPublisherConfig{
			Endpoint: cfg.EventPublisherEndpoint,
			Source:   cfg.EventPublisherSource,
			Timeout:  cfg.EventPublisherTimeout,
		})
	} else {
		eventPublisher = publisher.NewNoOpPublisher()
	}
	defer eventPublisher.Close()

	// Initialize OAuth provider
	var oauthProvider port.OAuthProvider
	if cfg.OAuthGoogleClientID != "" && cfg.OAuthGoogleClientSecret != "" {
		oauthProvider = security.NewGoogleOAuthAdapter(security.GoogleOAuthConfig{
			ClientID:     cfg.OAuthGoogleClientID,
			ClientSecret: cfg.OAuthGoogleClientSecret,
			RedirectURI:  cfg.OAuthRedirectURI,
		})
	}

	// Initialize services
	uidGen := security.NewUIDGenerator()
	userService := service.NewUserService(
		userRepo,
		profileRepo,
		pinRepo,
		deviceRepo,
		userDeviceRepo,
		passwordHasher,
		pinHasher,
		uidGen,
	)

	// Initialize auth service with all features
	authService := service.NewAuthService(
		userRepo,
		deviceRepo,
		userDeviceRepo,
		pinRepo,
		passwordHasher,
		pinHasher,
		tokenGen,
		uidGen,
		oauthProvider,
		tokenWhitelist,
		tokenBlacklist,
		eventPublisher,
	)

	// Initialize device service
	deviceService := service.NewDeviceService(
		deviceRepo,
		userDeviceRepo,
	)

	// Initialize user file service
	userFileRepo := repository.NewUserFileRepository(dbPool)
	userFileService := service.NewUserFileService(userFileRepo, userRepo, uidGen)

	// Initialize gRPC handlers
	userHandler := grpcadapter.NewUserHandler(userService)
	authHandler := grpcadapter.NewAuthHandler(authService)
	deviceHandler := grpcadapter.NewDeviceHandler(deviceService)
	userFileHandler := grpcadapter.NewUserFileHandler(userFileService)

	// Start gRPC server
	addr := fmt.Sprintf("%s:%d", cfg.App.IP, cfg.App.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("failed to listen", map[string]interface{}{
			"error": err.Error(),
		})
	}

	srv := grpc.NewServer()
	user.RegisterUserServiceServer(srv, userHandler)
	auth.RegisterAuthServiceServer(srv, authHandler)
	device.RegisterDeviceServiceServer(srv, deviceHandler)
	user_file.RegisterUserFileServiceServer(srv, userFileHandler)
	reflection.Register(srv)

	logger.Info("gRPC server listening", map[string]interface{}{
		"addr": addr,
	})

	// Handle shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down server", nil)
		srv.GracefulStop()
	}()

	if err := srv.Serve(lis); err != nil {
		logger.Fatal("failed to serve", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
