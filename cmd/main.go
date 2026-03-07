package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user-proto/gen/go/auth"
	"github.com/adityakw90/service-user-proto/gen/go/device"
	"github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user-proto/gen/go/user_file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcadapter "github.com/adityakw90/service-user/internal/adapter/api/grpc/handler"
	grpcMiddleware "github.com/adityakw90/service-user/internal/adapter/api/grpc/middleware"
	"github.com/adityakw90/service-user/internal/adapter/oauth"
	"github.com/adityakw90/service-user/internal/adapter/observer"
	"github.com/adityakw90/service-user/internal/adapter/publisher"
	"github.com/adityakw90/service-user/internal/adapter/repository"
	"github.com/adityakw90/service-user/internal/adapter/resolver"
	"github.com/adityakw90/service-user/internal/adapter/security"
	"github.com/adityakw90/service-user/internal/config"
	domainSignal "github.com/adityakw90/service-user/internal/core/domain/signal"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	portOAuth "github.com/adityakw90/service-user/internal/core/port/oauth"
	portobserver "github.com/adityakw90/service-user/internal/core/port/observer"
	"github.com/adityakw90/service-user/internal/core/service"
	"github.com/adityakw90/service-user/internal/infra"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	logger := infra.NewLogger()

	// initialize monitoring
	iMon, err := infra.NewMonitoring(&infra.MonitoringConfig{
		ServiceName:        cfg.Monitoring.ServiceName,
		Environment:        cfg.Monitoring.Environment,
		InstanceName:       cfg.Instance.Name,
		InstanceHost:       cfg.Instance.Host,
		LoggerLevel:        cfg.Monitoring.Logger.Level,
		TracerProvider:     cfg.Monitoring.Tracer.Provider,
		TracerProviderHost: cfg.Monitoring.Tracer.ProviderHost,
		TracerProviderPort: cfg.Monitoring.Tracer.ProviderPort,
		TracerSampleRatio:  cfg.Monitoring.Tracer.SampleRatio,
		TracerInsecure:     cfg.Monitoring.Tracer.Insecure,
		MetricProvider:     cfg.Monitoring.Metric.Provider,
		MetricProviderHost: cfg.Monitoring.Metric.ProviderHost,
		MetricProviderPort: cfg.Monitoring.Metric.ProviderPort,
		MetricInsecure:     cfg.Monitoring.Metric.Insecure,
	})
	if err != nil {
		logger.Fatal("failed to initialize monitoring", map[string]interface{}{
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
		QueryExecMode:         cfg.Database.QueryExecMode,
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

	// Connect to RabbitMQ using infra layer (if enabled)
	var rabbitmqConn *infra.RabbitMQConnection
	if cfg.EventPublisher.RabbitMQ.Enabled {
		rabbitmqConn, err = infra.NewRabbitMQConnection(ctx, infra.RabbitMQConfig{
			Host:                 cfg.Rabbit.Host,
			Port:                 cfg.Rabbit.Port,
			User:                 cfg.Rabbit.User,
			Password:             cfg.Rabbit.Password,
			Vhost:                cfg.Rabbit.Vhost,
			ReconnectInterval:    cfg.Rabbit.ReconnectInterval,
			ReconnectMaxAttempts: cfg.Rabbit.ReconnectMaxAttempts,
		}, iMon.Logger)
		if err != nil {
			logger.Fatal("failed to connect to rabbitmq", map[string]interface{}{
				"error": err.Error(),
			})
		}
		defer rabbitmqConn.Close()
		logger.Info("connected to rabbitmq", map[string]interface{}{
			"exchange": cfg.EventPublisher.RabbitMQ.Exchange,
		})
	}

	// Connect to Kafka using infra layer (if enabled)
	var kafkaConn *infra.KafkaConnection
	if cfg.EventPublisher.Kafka.Enabled {
		kafkaConn, err = infra.NewKafkaConnection(ctx, infra.KafkaConfig{
			Brokers:              cfg.Kafka.Brokers,
			MaxMessageBytes:      cfg.Kafka.MaxMessageBytes,
			Timeout:              time.Duration(cfg.Kafka.TimeoutSeconds) * time.Second,
			Compression:          cfg.Kafka.Compression,
			ReconnectInterval:    1 * time.Second,
			ReconnectMaxAttempts: 0, // 0 means infinite retries
		}, iMon.Logger)
		if err != nil {
			logger.Fatal("failed to connect to kafka", map[string]interface{}{
				"error": err.Error(),
			})
		}
		defer kafkaConn.Close()
		logger.Info("connected to kafka", map[string]interface{}{
			"brokers": cfg.Kafka.Brokers,
			"topic":   cfg.EventPublisher.Kafka.Topic,
		})
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool)
	profileRepo := repository.NewProfileRepository(dbPool)
	pinRepo := repository.NewPinRepository(dbPool)
	deviceRepo := repository.NewDeviceRepository(dbPool)
	userDeviceRepo := repository.NewUserDeviceRepository(dbPool)
	_ = repository.NewUserFileRepository(dbPool)

	// Initialize resolvers
	resolverProvider := resolver.NewResolverProvider(
		dbPool,
		redisClient,
		cfg.App.Code+":resolver",
		1*time.Hour,
		iMon.Logger,
		iMon.Tracer,
	)

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

	// Initialize security adapters
	securityAdapters, err := security.NewSecurityAdapters(ctx, security.SecurityConfig{
		LoginTracker: security.AttemptTrackerConfig{
			Backend:           cfg.Security.LoginTracker.Backend,
			LockoutThreshold:  cfg.Security.LoginTracker.LockoutThreshold,
			LockoutDuration:   cfg.Security.LoginTracker.LockoutDuration,
			LockoutCounterTTL: cfg.Security.LoginTracker.LockoutCounterTTL,
		},
		PINTracker: security.AttemptTrackerConfig{
			Backend:           cfg.Security.PINTracker.Backend,
			LockoutThreshold:  cfg.Security.PINTracker.LockoutThreshold,
			LockoutDuration:   cfg.Security.PINTracker.LockoutDuration,
			LockoutCounterTTL: cfg.Security.PINTracker.LockoutCounterTTL,
		},
		RateLimiter: security.RateLimiterConfig{
			Backend:    cfg.Security.RateLimiter.Backend,
			Limit:      cfg.Security.RateLimiter.Limit,
			WindowSize: cfg.Security.RateLimiter.WindowSize,
		},
	}, redisClient)
	if err != nil {
		logger.Fatal("failed to initialize security adapters", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer securityAdapters.Close()

	// --- Event Publishers Setup ---
	// Create a single unified event publisher for all event types

	var eventPublisher portEvent.EventPublisher
	if cfg.EventPublisher.Enabled {
		var backends []portEvent.EventPublisher

		// Redis Stream backend for all events
		if cfg.EventPublisher.Redis.Enabled && redisClient != nil {
			redisStreamPub, err := publisher.NewRedisPublisher(
				redisClient,
				publisher.RedisPublisherConfig{
					Stream: cfg.EventPublisher.Redis.Name,
					MaxLen: cfg.EventPublisher.Redis.MaxLen,
					Source: cfg.Instance.Name,
				},
				logger,
			)
			if err != nil {
				logger.Fatal("failed to create redis stream publisher", map[string]interface{}{
					"error": err.Error(),
				})
			}
			backends = append(backends, redisStreamPub)
		}

		// Kafka backend for all events
		// Use the existing connection if available (created earlier in main.go)
		if cfg.EventPublisher.Kafka.Enabled {
			if kafkaConn == nil {
				logger.Fatal("kafka connection is nil but kafka is enabled", nil)
			}
			kafkaPub := publisher.NewKafkaPublisherWithConn(
				kafkaConn,
				cfg.EventPublisher.Kafka.Topic,
				cfg.Instance.Name,
			)
			backends = append(backends, kafkaPub)
		}

		// RabbitMQ backend for all events
		// Use the existing connection if available (created earlier in main.go)
		if cfg.EventPublisher.RabbitMQ.Enabled {
			if rabbitmqConn == nil {
				logger.Fatal("rabbitmq connection is nil but rabbitmq is enabled", nil)
			}
			rabbitPub := publisher.NewRabbitMQPublisher(
				rabbitmqConn,
				publisher.RabbitMQPublisherConfig{
					Source:           cfg.Instance.Name,
					Exchange:         cfg.EventPublisher.RabbitMQ.Exchange,
					ExchangeType:     cfg.EventPublisher.RabbitMQ.ExchangeType,
					RoutingKeyPrefix: cfg.EventPublisher.RabbitMQ.RoutingKeyPrefix,
					Durable:          cfg.EventPublisher.RabbitMQ.Durable,
					ConfirmTimeout:   cfg.EventPublisher.RabbitMQ.ConfirmTimeout,
					MaxRetries:       cfg.EventPublisher.RabbitMQ.MaxRetries,
					RetryInterval:    cfg.EventPublisher.RabbitMQ.RetryInterval,
					QueueName:        cfg.EventPublisher.RabbitMQ.QueueName,
					QueueDurable:     cfg.EventPublisher.RabbitMQ.QueueDurable,
					QueueAutoDelete:  cfg.EventPublisher.RabbitMQ.QueueAutoDelete,
					QueueExclusive:   cfg.EventPublisher.RabbitMQ.QueueExclusive,
					QueueEnabled:     cfg.EventPublisher.RabbitMQ.QueueEnabled,
				},
			)
			// Setup infrastructure (exchange and optionally queue)
			if err := rabbitPub.SetupInfrastructure(); err != nil {
				logger.Fatal("failed to setup rabbitmq infrastructure", map[string]interface{}{
					"error":    err.Error(),
					"exchange": cfg.EventPublisher.RabbitMQ.Exchange,
				})
			}
			backends = append(backends, rabbitPub)
		}

		// HTTP backend for all events
		if cfg.EventPublisher.HTTP.Enabled {
			backends = append(backends, publisher.NewHTTPPublisher(
				publisher.HttpPublisherConfig{
					Endpoint: cfg.EventPublisher.HTTP.URL,
					Source:   cfg.Instance.Name,
					Timeout:  cfg.EventPublisher.HTTP.Timeout,
				},
			))
		}

		// Combine backends and wrap with async
		if len(backends) > 0 {
			multiBackend, err := publisher.NewMultiPublisher(logger, backends...)
			if err != nil {
				logger.Fatal("failed to create multi publisher", map[string]interface{}{
					"error": err.Error(),
				})
			}
			eventPublisher = publisher.NewAsyncPublisher(multiBackend, publisher.AsyncPublisherConfig{
				WorkerCount:   cfg.EventPublisher.WorkerCount,
				QueueSize:     cfg.EventPublisher.QueueSize,
				BatchSize:     cfg.EventPublisher.BatchSize,
				BatchTimeout:  cfg.EventPublisher.BatchTimeout,
				MaxRetries:    cfg.EventPublisher.MaxRetries,
				RetryInterval: cfg.EventPublisher.RetryInterval,
			})
		}
	}

	// Default to no-op if no publishers configured
	if eventPublisher == nil {
		eventPublisher = publisher.NewNoOpPublisher()
	}

	// Close publisher on shutdown
	defer eventPublisher.Close()

	// Initialize OAuth provider
	var oauthProvider portOAuth.OAuthProvider
	if cfg.OAuth.Enabled {
		oauthProvider, err = oauth.NewGoogleOAuth(&oauth.GoogleOAuthConfig{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
		}, redisClient)
		if err != nil {
			logger.Fatal("failed to initialize OAuth provider", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Create observers based on config
	authObserver := createAuthObserver(cfg, iMon.Logger, iMon.Tracer)
	userObserver := createUserObserver(cfg, iMon.Logger, iMon.Tracer)
	deviceObserver := createDeviceObserver(cfg, iMon.Logger, iMon.Tracer)
	userFileObserver := createUserFileObserver(cfg, iMon.Logger, iMon.Tracer)
	_ = createPinObserver(cfg, iMon.Logger, iMon.Tracer) // PIN operations handled in UserService

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
		tokenWhitelist,
		userObserver,
		eventPublisher,
		resolverProvider,
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
		authObserver,
		securityAdapters.LoginTracker,
		securityAdapters.RateLimiter,
	)

	// Initialize device service
	deviceService := service.NewDeviceService(
		deviceRepo,
		userDeviceRepo,
		deviceObserver,
		eventPublisher,
	)

	// Initialize user file service
	userFileRepo := repository.NewUserFileRepository(dbPool)
	userFileService := service.NewUserFileService(userFileRepo, userRepo, resolverProvider.User(), uidGen, userFileObserver, eventPublisher)

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

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryInterceptors(
				grpcMiddleware.UnaryRequestInterceptor(iMon),
			),
		),
		grpc.StreamInterceptor(
			grpcMiddleware.ChainStreamInterceptors(
				grpcMiddleware.StreamRequestInterceptor(iMon),
			),
		),
	)
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

func createAuthObserver(cfg *config.Config, logger gomon.Logger, tracer gomon.Tracer) portobserver.ServiceObserver[domainSignal.AuthSignal] {
	if cfg.Observer.Auth {
		return observer.NewAuthObserver(logger, tracer)
	}
	return observer.NewNoopObserver[domainSignal.AuthSignal]()
}

func createUserObserver(cfg *config.Config, logger gomon.Logger, tracer gomon.Tracer) portobserver.ServiceObserver[domainSignal.UserSignal] {
	if cfg.Observer.User {
		return observer.NewUserObserver(logger, tracer)
	}
	return observer.NewNoopObserver[domainSignal.UserSignal]()
}

func createDeviceObserver(cfg *config.Config, logger gomon.Logger, tracer gomon.Tracer) portobserver.ServiceObserver[domainSignal.DeviceSignal] {
	if cfg.Observer.Device {
		return observer.NewDeviceObserver(logger, tracer)
	}
	return observer.NewNoopObserver[domainSignal.DeviceSignal]()
}

func createUserFileObserver(cfg *config.Config, logger gomon.Logger, tracer gomon.Tracer) portobserver.ServiceObserver[domainSignal.UserFileSignal] {
	if cfg.Observer.UserFile {
		return observer.NewUserFileObserver(logger, tracer)
	}
	return observer.NewNoopObserver[domainSignal.UserFileSignal]()
}

func createPinObserver(cfg *config.Config, logger gomon.Logger, tracer gomon.Tracer) portobserver.ServiceObserver[domainSignal.PinSignal] {
	if cfg.Observer.Pin {
		return observer.NewPinObserver(logger, tracer)
	}
	return observer.NewNoopObserver[domainSignal.PinSignal]()
}
