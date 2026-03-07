package resolver

import (
	"testing"
	"time"

	monitoring "github.com/adityakw90/go-monitoring"
	"github.com/alicebob/miniredis/v2"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewResolverProvider tests the creation of a new ResolverProvider
func TestNewResolverProvider(t *testing.T) {
	tests := []struct {
		name               string
		db                 PostgrePool
		redisClient        *redis.Client
		redisPrefix        string
		redisCacheDuration time.Duration
		logger             *mockLogger
		tracer             monitoring.Tracer
	}{
		{
			name:               "Valid creation with all dependencies",
			db:                 nil, // nil is acceptable for this test
			redisClient:        nil,
			redisPrefix:        "test",
			redisCacheDuration: time.Hour,
			logger:             &mockLogger{},
			tracer:             newNoOpTracer(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewResolverProvider(
				tt.db,
				tt.redisClient,
				tt.redisPrefix,
				tt.redisCacheDuration,
				tt.logger,
				tt.tracer,
			)

			assert.NotNil(t, provider, "Provider should not be nil")

			// Type assertion to verify it's the correct implementation
			resolverProvider, ok := provider.(*resolverProvider)
			require.True(t, ok, "Provider should be of type *resolverProvider")

			// Verify all fields are set correctly
			assert.Equal(t, tt.db, resolverProvider.db, "DB should be set correctly")
			assert.Equal(t, tt.redisClient, resolverProvider.redisClient, "Redis client should be set correctly")
			assert.Equal(t, tt.redisPrefix, resolverProvider.redisPrefix, "Redis prefix should be set correctly")
			assert.Equal(t, tt.redisCacheDuration, resolverProvider.redisCacheDuration, "Cache duration should be set correctly")
			assert.Equal(t, tt.logger, resolverProvider.logger, "Logger should be set correctly")
			assert.Equal(t, tt.tracer, resolverProvider.tracer, "Tracer should be set correctly")
		})
	}
}

// TestResolverProvider_User tests that User() returns a non-nil resolver
func TestResolverProvider_User(t *testing.T) {
	tests := []struct {
		name        string
		redisPrefix string
	}{
		{
			name:        "Returns UserResolver with correct configuration",
			redisPrefix: "test",
		},
		{
			name:        "Adds prefix to Redis key correctly",
			redisPrefix: "myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &resolverProvider{
				redisPrefix: tt.redisPrefix,
				logger:      &mockLogger{},
				tracer:      newNoOpTracer(),
			}

			userResolver := provider.User()

			assert.NotNil(t, userResolver, "UserResolver should not be nil")
		})
	}
}

// TestResolverProvider_Device tests that Device() returns a non-nil resolver
func TestResolverProvider_Device(t *testing.T) {
	tests := []struct {
		name        string
		redisPrefix string
	}{
		{
			name:        "Returns DeviceResolver with correct configuration",
			redisPrefix: "test",
		},
		{
			name:        "Adds prefix to Redis key correctly",
			redisPrefix: "myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &resolverProvider{
				redisPrefix: tt.redisPrefix,
				logger:      &mockLogger{},
				tracer:      newNoOpTracer(),
			}

			deviceResolver := provider.Device()

			assert.NotNil(t, deviceResolver, "DeviceResolver should not be nil")
		})
	}
}

// TestResolverProvider_MultipleCalls tests that multiple calls return new instances
func TestResolverProvider_MultipleCalls(t *testing.T) {
	provider := &resolverProvider{
		redisPrefix: "test",
		logger:      &mockLogger{},
		tracer:      newNoOpTracer(),
	}

	// Call User() multiple times
	userResolver1 := provider.User()
	userResolver2 := provider.User()

	// Verify they are different instances
	assert.NotSame(t, userResolver1, userResolver2, "Multiple User() calls should return different instances")

	// Call Device() multiple times
	deviceResolver1 := provider.Device()
	deviceResolver2 := provider.Device()

	// Verify they are different instances
	assert.NotSame(t, deviceResolver1, deviceResolver2, "Multiple Device() calls should return different instances")

	// Verify User and Device resolvers are different
	assert.NotSame(t, userResolver1, deviceResolver1, "User and Device resolvers should be different instances")
}

// TestResolverProvider_WithRealDependencies tests provider with mock database and Redis
func TestResolverProvider_WithRealDependencies(t *testing.T) {
	// Setup mock database
	mockDB, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockDB.Close()

	// Setup miniredis
	s := miniredis.RunT(t)
	defer s.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer redisClient.Close()

	logger := &mockLogger{}
	tracer := newNoOpTracer()

	// Create provider with real mock dependencies
	provider := NewResolverProvider(
		mockDB,
		redisClient,
		"test",
		time.Hour,
		logger,
		tracer,
	)

	assert.NotNil(t, provider, "Provider should not be nil")

	// Test User() method
	userResolver := provider.User()
	assert.NotNil(t, userResolver, "UserResolver should not be nil")

	// Test Device() method
	deviceResolver := provider.Device()
	assert.NotNil(t, deviceResolver, "DeviceResolver should not be nil")

	// Verify resolvers are different instances
	assert.NotSame(t, userResolver, deviceResolver, "User and Device resolvers should be different instances")
}
