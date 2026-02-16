package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/adapter/oauth"
	"github.com/adityakw90/service-user/internal/adapter/observer"
	"github.com/adityakw90/service-user/internal/adapter/publisher"
	"github.com/adityakw90/service-user/internal/adapter/repository"
	"github.com/adityakw90/service-user/internal/adapter/resolver"
	"github.com/adityakw90/service-user/internal/adapter/security"
	"github.com/adityakw90/service-user/internal/config"
	portOAuth "github.com/adityakw90/service-user/internal/core/port/oauth"
	coreportsec "github.com/adityakw90/service-user/internal/core/port/security"
	coreportsvc "github.com/adityakw90/service-user/internal/core/port/service"
	svc "github.com/adityakw90/service-user/internal/core/service"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	golangOauth2 "golang.org/x/oauth2"
)

// TestServices holds all the services and dependencies for e2e testing.
type TestServices struct {
	Cfg             *config.Config
	DBPool          *pgxpool.Pool
	Redis           *redis.Client
	Hasher          coreportsec.Hasher
	PINHasher       coreportsec.Hasher
	TokenGen        coreportsec.TokenGenerator
	UserService     coreportsvc.UserService
	AuthService     coreportsvc.AuthService
	DeviceService   coreportsvc.DeviceService
	UserFileService coreportsvc.UserFileService
}

// SetupTestServices initializes all services for e2e testing.
// It connects to the database and Redis, initializes repositories,
// and creates service instances with test configuration.
func SetupTestServices(t *testing.T, ctx context.Context) (*TestServices, error) {
	t.Helper()

	// Load configuration
	cfg, err := LoadTestConfig(t)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// temporaty initialzie the monitoring here
	monitoring, err := infra.NewMonitoring(&infra.MonitoringConfig{
		ServiceName:         "test-service",
		Environment:         "development",
		InstanceName:        "test-instance",
		InstanceHost:        "test-host",
		LoggerLevel:         "error",
		LoggerCallerSkipNum: 1,
		TracerProvider:      "stdout",
		TracerProviderHost:  "localhost",
		TracerProviderPort:  6831,
		TracerSampleRatio:   1.0,
		MetricProvider:      "stdout",
		MetricProviderHost:  "localhost",
		MetricProviderPort:  9090,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize monitoring: %w", err)
	}

	// Connect to database
	dbPool, err := NewTestPostgreConnection(t, ctx, cfg)
	if err != nil {
		t.Fatal(err)
		return nil, err
	}
	// dbPool, err := infra.NewPostgreConnection(ctx, &infra.PostgreConfig{
	// 	Host:                  cfg.Database.Host,
	// 	Port:                  cfg.Database.Port,
	// 	User:                  cfg.Database.User,
	// 	Password:              cfg.Database.Password,
	// 	Name:                  cfg.Database.Name,
	// 	SslMode:               cfg.Database.SslMode,
	// 	Timezone:              cfg.Database.Timezone,
	// 	MinConns:              cfg.Database.MinConns,
	// 	MinIdleConns:          cfg.Database.MinIdleConns,
	// 	MaxConns:              cfg.Database.MaxConns,
	// 	MaxConnIdleTime:       cfg.Database.MaxConnIdleTime,
	// 	MaxConnLifetime:       cfg.Database.MaxConnLifetime,
	// 	MaxConnLifetimeJitter: cfg.Database.MaxConnLifetimeJitter,
	// 	HealthCheckPeriod:     cfg.Database.HealthCheckPeriod,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to connect to database: %w", err)
	// }

	// Connect to Redis
	redisClient, err := NewTestRedisConnection(t, ctx, cfg)
	if err != nil {
		t.Fatal(err)
		return nil, err
	}
	// redisClient, err := infra.NewRedisConnection(context.Background(), &infra.RedisConfig{
	// 	Host:              cfg.Redis.Host,
	// 	Port:              cfg.Redis.Port,
	// 	User:              cfg.Redis.User,
	// 	Password:          cfg.Redis.Password,
	// 	DB:                cfg.Redis.DB,
	// 	PoolSize:          cfg.Redis.PoolSize,
	// 	PoolTimeout:       cfg.Redis.PoolTimeout,
	// 	ConnectionIdleMin: cfg.Redis.ConnectionIdleMin,
	// })
	// if err != nil {
	// 	dbPool.Close()
	// 	return nil, fmt.Errorf("failed to connect to redis: %w", err)
	// }

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool)
	profileRepo := repository.NewProfileRepository(dbPool)
	pinRepo := repository.NewPinRepository(dbPool)
	deviceRepo := repository.NewDeviceRepository(dbPool)
	userDeviceRepo := repository.NewUserDeviceRepository(dbPool)

	// Initialize resolver
	userResolver := resolver.NewUserResolver(
		dbPool,
		redisClient,
		"SUS:resolver-test:user",
		15*time.Minute,
		monitoring.Logger,
		monitoring.Tracer,
	)

	// Initialize hashers (use fast settings for tests)
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
		return nil, fmt.Errorf("failed to initialize password hasher: %w", err)
	}

	pinHasher, err := security.NewHasher(cfg.PINHasher.Type, map[string]any{
		"cost":        cfg.PINHasher.Cost,
		"salt":        cfg.PINHasher.Salt,
		"memory":      cfg.PINHasher.Memory,
		"iterations":  cfg.PINHasher.Iterations,
		"parallelism": cfg.PINHasher.Parallelism,
		"saltLength":  cfg.PINHasher.SaltLength,
		"keyLength":   cfg.PINHasher.KeyLength,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PIN hasher: %w", err)
	}

	// Initialize token generator
	tokenGen := security.NewJWTGenerator(
		cfg.Jwt.SecretKey,
		cfg.Jwt.AccessExpiry,
		cfg.Jwt.RefreshExpiry,
	)

	// Initialize token blacklist/whitelist
	tokenBlacklist := security.NewTokenBlacklistAdapter(redisClient, "test-token-blacklist:", 24*time.Hour)
	tokenWhitelist := security.NewTokenWhitelistAdapter(redisClient, "test-token-whitelist:", 15*time.Minute)

	// Create event publishers (no-op for tests)
	eventPublisher := publisher.NewNoOpPublisher()

	// Initialize UID generator
	uidGen := security.NewUIDGenerator()

	// Initialize Observer
	authObserver := observer.NewAuthObserver(monitoring.Logger, monitoring.Tracer)
	userObserver := observer.NewUserObserver(monitoring.Logger, monitoring.Tracer)
	deviceObserver := observer.NewDeviceObserver(monitoring.Logger, monitoring.Tracer)
	userFileObserver := observer.NewUserFileObserver(monitoring.Logger, monitoring.Tracer)

	// Initialize services
	userService := svc.NewUserService(
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
	)

	// Initialize OAuth provider if configured
	var oauthProvider portOAuth.OAuthProvider
	if cfg.OAuth.Enabled && cfg.OAuth.Google.Enabled {
		oauthConfig := &oauth.GoogleOAuthConfig{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			Scopes:       []string{"openid", "email", "profile"},
		}
		// Set custom endpoint if TokenURL is provided (for testing with mock servers)
		if cfg.OAuth.Google.TokenURL != "" {
			authURL := cfg.OAuth.Google.AuthURL
			if authURL == "" {
				// Use default Google Auth URL if not specified
				authURL = "https://accounts.google.com/o/oauth2/auth"
			}
			oauthConfig.Endpoint = &golangOauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: cfg.OAuth.Google.TokenURL,
			}
		}
		// Set UserInfoURL if provided (for testing with mock servers)
		if cfg.OAuth.Google.UserInfoURL != "" {
			oauthConfig.UserInfoURL = cfg.OAuth.Google.UserInfoURL
		}
		oauthProvider, err = oauth.NewGoogleOAuth(oauthConfig, redisClient)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize OAuth provider: %w", err)
		}
	}

	authService := svc.NewAuthService(
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
		security.NewNoopAttemptTracker(),
		security.NewNoopRateLimiter(),
	)

	// Initialize device service
	deviceService := svc.NewDeviceService(
		deviceRepo,
		userDeviceRepo,
		deviceObserver,
		eventPublisher,
	)

	// Initialize user file service with resolver
	userFileRepo := repository.NewUserFileRepository(dbPool)
	userFileService := svc.NewUserFileService(userFileRepo, userRepo, userResolver, uidGen, userFileObserver, eventPublisher)

	// prepare db and redis
	// TruncateTestTables(t, ctx, dbPool)
	// TruncateTestRedis(t, ctx, redisClient)

	// cleanup func
	t.Cleanup(func() {
		dbPool.Close()
		redisClient.Close()
	})

	return &TestServices{
		Cfg:             cfg,
		DBPool:          dbPool,
		Redis:           redisClient,
		Hasher:          passwordHasher,
		PINHasher:       pinHasher,
		TokenGen:        tokenGen,
		UserService:     userService,
		AuthService:     authService,
		DeviceService:   deviceService,
		UserFileService: userFileService,
	}, nil
}
