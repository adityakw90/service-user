package util

import (
	"context"
	"fmt"
	"time"

	"github.com/adityakw90/service-user/internal/adapter/publisher"
	"github.com/adityakw90/service-user/internal/adapter/repository"
	"github.com/adityakw90/service-user/internal/adapter/security"
	"github.com/adityakw90/service-user/internal/config"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
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

// SetupTestServices initializes all services for e2e testing.
// It connects to the database and Redis, initializes repositories,
// and creates service instances with test configuration.
func SetupTestServices(ctx context.Context) (*TestServices, error) {
	// Load configuration
	cfg, err := LoadTestConfig()
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

	// Initialize user file service
	userFileRepo := repository.NewUserFileRepository(dbPool)
	userFileService := svc.NewUserFileService(userFileRepo, userRepo, uidGen)

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

// TeardownTestServices closes all connections.
func TeardownTestServices(svc *TestServices) {
	if svc.DBPool != nil {
		svc.DBPool.Close()
	}
	if svc.Redis != nil {
		svc.Redis.Close()
	}
}

// SetupTest runs before each test.
func (s *TestServices) SetupTest() {
	// Clean up test data before each test
	ctx := context.Background()
	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user_device"`)
	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "device"`)
	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user_pin"`)
	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user_profile"`)
	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user_file"`)
	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user"`)

	// Clean up Redis keys
	iter := s.Redis.Scan(ctx, 0, "test:*", 100).Iterator()
	for iter.Next(ctx) {
		s.Redis.Del(ctx, iter.Val())
	}
}

// CreateTestUserViaService creates a test user through the service layer.
// This is the recommended way to create users in tests as it exercises
// the full service logic including hashing.
func CreateTestUserViaService(ctx context.Context, svc *TestServices, username, email, password string) (*TestUser, error) {
	user, err := svc.UserService.Create(ctx, &params.UserCreateParam{
		Username: username,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create test user: %w", err)
	}

	return &TestUser{
		User:     user,
		Password: password,
	}, nil
}

// TestUser wraps a User with the plain-text password for testing.
type TestUser struct {
	User     *model.User
	Password string
}

// CleanupTestData cleans up all test data from database and Redis.
func CleanupTestData(ctx context.Context, svc *TestServices) error {
	// Clean up database
	if err := SetupTestDatabase(ctx, svc.DBPool); err != nil {
		return fmt.Errorf("failed to cleanup database: %w", err)
	}

	// Clean up Redis
	if err := SetupTestRedis(ctx, svc.Redis); err != nil {
		return fmt.Errorf("failed to cleanup Redis: %w", err)
	}

	return nil
}

// WaitForInfrastructure waits for both database and Redis to be ready.
func WaitForInfrastructure(ctx context.Context, dbURL, redisURL string, maxAttempts int) error {
	// Wait for database
	if err := WaitForDatabase(ctx, dbURL, maxAttempts); err != nil {
		return fmt.Errorf("database not ready: %w", err)
	}

	// Wait for Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
		DB:   0,
	})
	defer redisClient.Close()

	if err := WaitForRedis(ctx, redisClient, maxAttempts); err != nil {
		return fmt.Errorf("redis not ready: %w", err)
	}

	return nil
}
