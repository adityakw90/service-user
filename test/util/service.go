package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/adapter/event"
	"github.com/adityakw90/service-user/internal/adapter/executor"
	"github.com/adityakw90/service-user/internal/adapter/oauth"
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
	"golang.org/x/oauth2"
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
		TracerSampleRatio:   0,
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

	// Connect to Redis
	redisClient, err := NewTestRedisConnection(t, ctx, cfg)
	if err != nil {
		t.Fatal(err)
		return nil, err
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool, monitoring.Tracer, monitoring.Logger)
	profileRepo := repository.NewProfileRepository(dbPool, monitoring.Tracer, monitoring.Logger)
	pinRepo := repository.NewPinRepository(dbPool, monitoring.Tracer, monitoring.Logger)
	deviceRepo := repository.NewDeviceRepository(dbPool, monitoring.Tracer, monitoring.Logger)
	userDeviceRepo := repository.NewUserDeviceRepository(dbPool, monitoring.Tracer, monitoring.Logger)

	// Initialize resolver
	resolverProvider := resolver.NewResolverProvider(
		dbPool,
		redisClient,
		"SUS:resolver-test",
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
	tokenBlacklist := security.NewTokenBlacklistAdapter(redisClient, "test-token-blacklist:", 24*time.Hour, monitoring.Tracer, monitoring.Logger)
	tokenWhitelist := security.NewTokenWhitelistAdapter(redisClient, "test-token-whitelist:", 15*time.Minute, monitoring.Tracer, monitoring.Logger)

	// Create event publishers (no-op for tests)
	eventPublisher := event.NewNoOpPublisher()

	// Initialize UID generator
	uidGen := security.NewUIDGenerator()

	// Initialize Executor
	serviceExecutor := executor.NewServiceExecutor(monitoring.Logger, monitoring.Tracer)
	t.Cleanup(serviceExecutor.Close)

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
		eventPublisher,
		resolverProvider,
	)

	// Initialize OAuth provider if configured
	var oauthProvider portOAuth.OAuthProvider
	if cfg.OAuth.Enabled && cfg.OAuth.Google.Enabled {
		oauthConfig := &oauth.GoogleOAuthConfig{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			Scopes:       []string{"openid", "email", "profile"},
		}

		// Check for custom OAuth endpoints (for testing with mock servers)
		tokenURL := os.Getenv("OAUTH_GOOGLE_TOKEN_URL")
		userInfoURL := os.Getenv("OAUTH_GOOGLE_USER_INFO_URL")
		authURL := os.Getenv("OAUTH_GOOGLE_AUTH_URL")

		if userInfoURL != "" {
			oauthConfig.UserInfoURL = userInfoURL
		}
		if tokenURL != "" {
			if authURL == "" {
				// Default to Google's auth URL if only token URL is provided
				authURL = "https://accounts.google.com/o/oauth2/auth"
			}
			oauthConfig.Endpoint = &oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			}
		}

		oauthProvider, err = oauth.NewGoogleOAuth(oauthConfig, redisClient, monitoring.Tracer, monitoring.Logger)
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
		serviceExecutor,
		eventPublisher,
		security.NewNoopAttemptTracker(),
		security.NewNoopRateLimiter(),
	)

	// Initialize device service
	deviceService := svc.NewDeviceService(
		deviceRepo,
		userDeviceRepo,
		eventPublisher,
	)

	// Initialize user file service with resolver
	userFileRepo := repository.NewUserFileRepository(dbPool, monitoring.Tracer, monitoring.Logger)
	userFileService := svc.NewUserFileService(userFileRepo, userRepo, resolverProvider.User(), uidGen, eventPublisher)

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
