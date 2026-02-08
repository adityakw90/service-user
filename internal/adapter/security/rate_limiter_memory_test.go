package security

import (
	"context"
	"testing"
	"time"
)

func TestMemoryRateLimiter_Acquire(t *testing.T) {
	limit := 5
	windowSize := 60 * time.Second
	limiter := NewMemoryRateLimiter(limit, windowSize)

	ctx := context.Background()
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

func TestMemoryRateLimiter_Reset(t *testing.T) {
	limit := 5
	windowSize := 60 * time.Second
	limiter := NewMemoryRateLimiter(limit, windowSize)

	ctx := context.Background()
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

func TestMemoryRateLimiter_SlidingWindow(t *testing.T) {
	limit := 5
	windowSize := 1 * time.Second
	limiter := NewMemoryRateLimiter(limit, windowSize)

	ctx := context.Background()
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

func TestMemoryRateLimiter_MultipleIPs(t *testing.T) {
	limit := 3
	windowSize := 60 * time.Second
	limiter := NewMemoryRateLimiter(limit, windowSize)

	ctx := context.Background()
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

func TestMemoryRateLimiter_Cleanup(t *testing.T) {
	limit := 10
	windowSize := 1 * time.Second
	limiter := NewMemoryRateLimiter(limit, windowSize)

	ctx := context.Background()
	deviceIp := "192.168.1.1"

	// Make some requests
	for i := 0; i < 5; i++ {
		limiter.Acquire(ctx, deviceIp)
	}

	// Run cleanup (cast to concrete type to access Cleanup method)
	ml := limiter.(*MemoryRateLimiter)
	ml.Cleanup()

	// Cleanup should have cleared old entries
	// We can't directly test this, but the test ensures no panic occurs
}
