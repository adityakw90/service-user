package security

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/service-user/test/util/redis"
)

func TestRedisRateLimiter_Acquire(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	limit := 5
	windowSize := 60 * time.Second
	limiter := NewRedisRateLimiter(client, limit, windowSize)
	deviceIp := "192.168.1.1"

	tests := []struct {
		name        string
		setup       func()
		requests    int
		wantAllowed bool
	}{
		{
			name:        "First request allowed",
			requests:    1,
			wantAllowed: true,
		},
		{
			name:        "Multiple requests within limit",
			setup:       func() { limiter.Reset(ctx, deviceIp) },
			requests:    3,
			wantAllowed: true,
		},
		{
			name:        "Request at limit allowed",
			setup:       func() { limiter.Reset(ctx, deviceIp) },
			requests:    5,
			wantAllowed: true,
		},
		{
			name:        "Request over limit denied",
			setup:       func() { limiter.Reset(ctx, deviceIp) },
			requests:    6,
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			var allowed bool
			var err error

			for i := 0; i < tt.requests; i++ {
				allowed, err = limiter.Acquire(ctx, deviceIp)
			}

			if err != nil {
				t.Fatalf("Acquire() unexpected error = %v", err)
			}
			if allowed != tt.wantAllowed {
				t.Errorf("Acquire() allowed = %v, want %v", allowed, tt.wantAllowed)
			}
		})
	}
}

func TestRedisRateLimiter_Reset(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	limit := 5
	windowSize := 60 * time.Second
	limiter := NewRedisRateLimiter(client, limit, windowSize)
	deviceIp := "192.168.1.1"

	// Use up the limit
	for i := 0; i < limit; i++ {
		allowed, _ := limiter.Acquire(ctx, deviceIp)
		if !allowed {
			t.Fatal("Expected all initial requests to be allowed")
		}
	}

	// Should be at limit
	allowed, _ := limiter.Acquire(ctx, deviceIp)
	if allowed {
		t.Error("Expected request over limit to be denied")
	}

	// Reset
	err := limiter.Reset(ctx, deviceIp)
	if err != nil {
		t.Fatalf("Reset() unexpected error = %v", err)
	}

	// Should be allowed again
	allowed, _ = limiter.Acquire(ctx, deviceIp)
	if !allowed {
		t.Error("Expected request after reset to be allowed")
	}
}

func TestRedisRateLimiter_SlidingWindow(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	limit := 5
	windowSize := 1 * time.Second
	limiter := NewRedisRateLimiter(client, limit, windowSize)
	deviceIp := "192.168.1.1"

	// Use up the limit
	for i := 0; i < limit; i++ {
		allowed, _ := limiter.Acquire(ctx, deviceIp)
		if !allowed {
			t.Fatal("Expected all initial requests to be allowed")
		}
	}

	// Should be at limit
	allowed, _ := limiter.Acquire(ctx, deviceIp)
	if allowed {
		t.Error("Expected request at limit to be denied")
	}

	// Wait for window to pass
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again
	allowed, _ = limiter.Acquire(ctx, deviceIp)
	if !allowed {
		t.Error("Expected request after window expiry to be allowed")
	}
}

func TestRedisRateLimiter_MultipleIPs(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	limit := 3
	windowSize := 60 * time.Second
	limiter := NewRedisRateLimiter(client, limit, windowSize)

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Use up limit for ip1
	for i := 0; i < limit; i++ {
		limiter.Acquire(ctx, ip1)
	}

	// ip1 should be at limit
	allowed, _ := limiter.Acquire(ctx, ip1)
	if allowed {
		t.Error("Expected ip1 to be at limit")
	}

	// ip2 should still be allowed
	allowed, _ = limiter.Acquire(ctx, ip2)
	if !allowed {
		t.Error("Expected ip2 request to be allowed")
	}
}

func TestRedisRateLimiter_ConcurrentRequests(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	limit := 100
	windowSize := 60 * time.Second
	limiter := NewRedisRateLimiter(client, limit, windowSize)
	deviceIp := "192.168.1.1"

	// Launch concurrent requests
	done := make(chan bool, limit)
	for i := 0; i < limit; i++ {
		go func() {
			allowed, err := limiter.Acquire(ctx, deviceIp)
			if err != nil {
				t.Errorf("Acquire() unexpected error = %v", err)
			}
			if !allowed {
				t.Errorf("Acquire() should be allowed for all %d requests", limit)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < limit; i++ {
		<-done
	}
}
