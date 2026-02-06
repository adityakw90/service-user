package util

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/adapter/publisher"
	"github.com/adityakw90/service-user/internal/adapter/repository"
	"github.com/adityakw90/service-user/internal/adapter/security"
	"github.com/adityakw90/service-user/internal/config"
	coreportrepo "github.com/adityakw90/service-user/internal/core/port/repository"
	coreportresolver "github.com/adityakw90/service-user/internal/core/port/resolver"
	coreportsec "github.com/adityakw90/service-user/internal/core/port/security"
	coreportsvc "github.com/adityakw90/service-user/internal/core/port/service"
	svc "github.com/adityakw90/service-user/internal/core/service"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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

// testUserResolver is a simple resolver implementation for tests that wraps the user repository.
type testUserResolver struct {
	userRepo coreportrepo.UserRepository
}

// newTestUserResolver creates a new test user resolver.
func newTestUserResolver(userRepo coreportrepo.UserRepository) coreportresolver.UserResolver {
	return &testUserResolver{userRepo: userRepo}
}

func (r *testUserResolver) IDsByUIDs(ctx context.Context, userUIDs []string) (map[string]int64, error) {
	result := make(map[string]int64)
	for _, uid := range userUIDs {
		user, err := r.userRepo.GetByUID(ctx, uid)
		if err != nil {
			return nil, err
		}
		result[uid] = user.ID
	}
	return result, nil
}

func (r *testUserResolver) UIDsByIDs(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	for _, id := range userIDs {
		user, err := r.userRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		result[id] = user.UID
	}
	return result, nil
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

	// Connect to database
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
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Connect to Redis
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
		dbPool.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool)
	profileRepo := repository.NewProfileRepository(dbPool)
	pinRepo := repository.NewPinRepository(dbPool)
	deviceRepo := repository.NewDeviceRepository(dbPool)
	userDeviceRepo := repository.NewUserDeviceRepository(dbPool)

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

	// Create event publisher (no-op for tests)
	eventPublisher := publisher.NewNoOpPublisher()

	// Initialize UID generator
	uidGen := security.NewUIDGenerator()

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
	)

	authService := svc.NewAuthService(
		userRepo,
		deviceRepo,
		userDeviceRepo,
		pinRepo,
		passwordHasher,
		pinHasher,
		tokenGen,
		uidGen,
		nil, // oauthProvider
		tokenWhitelist,
		tokenBlacklist,
		eventPublisher,
	)

	// Initialize device service
	deviceService := svc.NewDeviceService(
		deviceRepo,
		userDeviceRepo,
	)

	// Initialize user file service with resolver
	userFileRepo := repository.NewUserFileRepository(dbPool)
	userResolver := newTestUserResolver(userRepo)
	userFileService := svc.NewUserFileService(userFileRepo, userRepo, userResolver, uidGen)

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
