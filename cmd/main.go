package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	grpcadapter "github.com/adityakw90/service-user/internal/adapter/api/grpc"
	"github.com/adityakw90/service-user/internal/adapter/event"
	"github.com/adityakw90/service-user/internal/adapter/executor"
	"github.com/adityakw90/service-user/internal/adapter/oauth"
	"github.com/adityakw90/service-user/internal/adapter/observer"
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
	// Define all command-line flags
	config.InitFlags()

	// Handle --version flag early (before loading config)
	handleVersionFlag()

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
	logger.Info("connected to database", map[string]interface{}{
		"host": cfg.Database.Host,
		"port": cfg.Database.Port,
	})

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
	logger.Info("connected to redis", map[string]interface{}{
		"host": cfg.Redis.Host,
		"port": cfg.Redis.Port,
	})

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
	var eventPublisher portEvent.EventPublisher
	if cfg.EventPublisher.Enabled {
		var backendsPublisher []portEvent.EventPublisher

		// setup http publisher
		if cfg.EventPublisher.HTTP.Enabled {
			eventHttpPublisher := event.NewHTTPPublisher(
				cfg.EventPublisher.HTTP.URL,
				cfg.EventPublisher.HTTP.Timeout,
				iMon.Logger,
				iMon.Tracer,
			)
			backendsPublisher = append(backendsPublisher, eventHttpPublisher)
		}

		// setup rabbitmq publisher
		if cfg.EventPublisher.RabbitMQ.Enabled {
			var rabbitmqConn *infra.Rabbit
			rabbitmqConn, err = infra.NewRabbitConnection(ctx, infra.RabbitConfig{
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
				"host":  cfg.Rabbit.Host,
				"port":  cfg.Rabbit.Port,
				"user":  cfg.Rabbit.User,
				"vhost": cfg.Rabbit.Vhost,
			})

			eventRabbitPublisher := event.NewRabbitmqPublisher(
				rabbitmqConn,
				event.RabbitmqPublisherConfig{
					Exchange:         cfg.EventPublisher.RabbitMQ.Exchange,
					RoutingKeyPrefix: cfg.EventPublisher.RabbitMQ.RoutingKeyPrefix,
					ConfirmTimeout:   cfg.EventPublisher.RabbitMQ.ConfirmTimeout,
					MaxRetries:       cfg.EventPublisher.RabbitMQ.MaxRetries,
					RetryInterval:    cfg.EventPublisher.RabbitMQ.RetryInterval,
				},
				iMon.Logger,
				iMon.Tracer,
			)
			backendsPublisher = append(backendsPublisher, eventRabbitPublisher)
		}

		if len(backendsPublisher) == 1 {
			eventPublisher = backendsPublisher[0]
		} else if len(backendsPublisher) > 1 {
			eventPublisher = event.NewMultiEventPublisher(
				iMon.Logger,
				iMon.Tracer,
				backendsPublisher...,
			)
		}
	}

	// Default to no-op if no publishers configured
	if eventPublisher == nil {
		eventPublisher = event.NewNoOpPublisher()
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

	// create executor
	exc := executor.NewServiceExecutor(iMon.Logger, iMon.Tracer)
	defer exc.Close()

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
		exc,
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

	// Create gRPC server with centralized setup
	srv := grpcadapter.NewServer(
		authService,
		userService,
		deviceService,
		userFileService,
		iMon,
	)

	// Start gRPC server in background
	addr := fmt.Sprintf("%s:%d", cfg.App.IP, cfg.App.Port)
	// Handle shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down server", nil)
		srv.Stop()
	}()

	if err := srv.Start(ctx, addr); err != nil {
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
