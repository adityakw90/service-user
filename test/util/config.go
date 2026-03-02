package testutil

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/config"
)

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func LoadTestConfig(t *testing.T) (*config.Config, error) {
	t.Helper()

	dbPort, err := strconv.Atoi(getEnv("DATABASE_PORT", "5432"))
	if err != nil {
		return nil, err
	}
	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, err
	}
	rabbitPort, err := strconv.Atoi(getEnv("RABBITMQ_PORT", "5672"))
	if err != nil {
		return nil, err
	}
	return &config.Config{
		Database: config.DatabaseConfig{
			Host:                  getEnv("DATABASE_HOST", "localhost"),
			Port:                  dbPort,
			User:                  getEnv("DATABASE_USER", "postgres"),
			Password:              getEnv("DATABASE_PASSWORD", "postgres"),
			Name:                  getEnv("DATABASE_NAME", "service_db"),
			SslMode:               getEnv("DATABASE_SSL_MODE", "disable"),
			Timezone:              "UTC",
			MinConns:              1,
			MinIdleConns:          1,
			MaxConns:              10,
			MaxConnIdleTime:       10 * time.Minute,
			MaxConnLifetime:       30 * time.Minute,
			MaxConnLifetimeJitter: 5 * time.Minute,
			HealthCheckPeriod:     1 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host:              getEnv("REDIS_HOST", "localhost"),
			Port:              redisPort,
			User:              getEnv("REDIS_USER", ""),
			Password:          getEnv("REDIS_PASSWORD", ""),
			DB:                0,
			PoolSize:          10,
			PoolTimeout:       5 * time.Second,
			ConnectionIdleMin: 1,
		},
		Rabbit: config.RabbitConfig{
			Host:                 getEnv("RABBITMQ_HOST", "localhost"),
			Port:                 rabbitPort,
			User:                 getEnv("RABBITMQ_USER", "rabbit"),
			Password:             getEnv("RABBITMQ_PASSWORD", "password"),
			Vhost:                getEnv("RABBITMQ_VHOST", "/"),
			ReconnectInterval:    1 * time.Second,
			ReconnectMaxAttempts: 0, // 0 means infinite retries
		},
		Observer: config.ObserverConfig{
			Auth:     true,
			User:     true,
			Device:   true,
			UserFile: true,
			Pin:      true,
		},
		Security: config.SecurityConfig{
			LoginTracker: config.AttemptTrackerConfig{
				Backend:           "redis",
				LockoutCounterTTL: 30 * time.Minute,
				LockoutDuration:   15 * time.Minute,
				LockoutThreshold:  5,
			},
			PINTracker: config.AttemptTrackerConfig{
				Backend:           "redis",
				LockoutCounterTTL: 30 * time.Minute,
				LockoutDuration:   15 * time.Minute,
				LockoutThreshold:  3,
			},
			RateLimiter: config.RateLimiterConfig{
				Backend:    "redis",
				Limit:      100,
				WindowSize: time.Hour,
			},
		},
		Jwt: config.JWTConfig{
			SecretKey:     "test-secret-key",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 24 * time.Hour,
		},
		PasswordHasher: config.HasherConfig{
			Type:        "bcrypt",
			Cost:        10,
			Salt:        "",
			Memory:      64 * 1024,
			Iterations:  2,
			Parallelism: 1,
			SaltLength:  16,
			KeyLength:   32,
		},
		PINHasher: config.HasherConfig{
			Type:        "bcrypt",
			Cost:        10,
			Salt:        "",
			Memory:      64 * 1024,
			Iterations:  2,
			Parallelism: 1,
			SaltLength:  16,
			KeyLength:   32,
		},
		EventPublisher: config.EventPublisherConfig{
			HTTP: config.PublisherHTTPConfig{
				Enabled: true,
				URL:     "http://localhost:8080",
				Timeout: 5 * time.Second,
			},
			Redis: config.PublisherRedisConfig{
				Enabled: true,
				Name:    "test_service_user",
			},
			RabbitMQ: config.PublisherRabbitMQConfig{
				Enabled:          true,
				Exchange:         "test_service_user",
				ExchangeType:     "topic",
				RoutingKeyPrefix: "test_service_user.",
				Durable:          true,
			},
		},
		OAuth: config.OauthConfig{
			Enabled: true,
			Google: config.OauthGoogleConfig{
				Enabled:      true,
				ClientID:     getEnv("OAUTH_GOOGLE_CLIENT_ID", "test-client-id"),
				ClientSecret: getEnv("OAUTH_GOOGLE_CLIENT_SECRET", "test-client-secret"),
			},
		},
	}, nil
}
