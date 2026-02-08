package security

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/service-user/test/util/redis"
)

func TestRedisAttemptTracker_Track(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewRedisAttemptTracker(client, threshold, lockoutDuration, counterTTL)
	userUID := "user-123"

	tests := []struct {
		name        string
		setup       func()
		attempts    int
		wantLocked  bool
		wantErr     bool
	}{
		{
			name:     "First attempt does not lock",
			attempts: 1,
			wantLocked: false,
			wantErr:   false,
		},
		{
			name:     "Second attempt does not lock",
			attempts: 2,
			wantLocked: false,
			wantErr:   false,
		},
		{
			name:     "Third attempt locks account",
			attempts: 3,
			wantLocked: true,
			wantErr:   true,
		},
		{
			name: "Fourth attempt already locked",
			setup: func() {
				// Use up all attempts first
				for i := 0; i < threshold; i++ {
					tracker.Track(ctx, userUID)
				}
			},
			attempts:   1,
			wantLocked: true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			var err error
			for i := 0; i < tt.attempts; i++ {
				err = tracker.Track(ctx, userUID)
			}

			if tt.wantErr && err == nil {
				t.Error("Track() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Track() unexpected error = %v", err)
			}

			locked, _ := tracker.IsLocked(ctx, userUID)
			if locked != tt.wantLocked {
				t.Errorf("IsLocked() = %v, want %v", locked, tt.wantLocked)
			}

			// Reset for next test
			tracker.Reset(ctx, userUID)
		})
	}
}

func TestRedisAttemptTracker_IsLocked(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewRedisAttemptTracker(client, threshold, lockoutDuration, counterTTL)
	userUID := "user-123"

	// Initially not locked
	locked, err := tracker.IsLocked(ctx, userUID)
	if err != nil {
		t.Fatalf("IsLocked() unexpected error = %v", err)
	}
	if locked {
		t.Error("IsLocked() = true, want false (initial state)")
	}

	// Track attempts until lockout
	for i := 0; i < threshold; i++ {
		tracker.Track(ctx, userUID)
	}

	// Should be locked now
	locked, err = tracker.IsLocked(ctx, userUID)
	if err != nil {
		t.Fatalf("IsLocked() unexpected error = %v", err)
	}
	if !locked {
		t.Error("IsLocked() = false, want true (after threshold attempts)")
	}
}

func TestRedisAttemptTracker_Reset(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewRedisAttemptTracker(client, threshold, lockoutDuration, counterTTL)
	userUID := "user-123"

	// Track attempts until lockout
	for i := 0; i < threshold; i++ {
		tracker.Track(ctx, userUID)
	}

	// Should be locked
	locked, _ := tracker.IsLocked(ctx, userUID)
	if !locked {
		t.Fatal("IsLocked() = false, want true (after threshold attempts)")
	}

	// Reset
	err := tracker.Reset(ctx, userUID)
	if err != nil {
		t.Fatalf("Reset() unexpected error = %v", err)
	}

	// Should no longer be locked
	locked, _ = tracker.IsLocked(ctx, userUID)
	if locked {
		t.Error("IsLocked() = true, want false (after reset)")
	}

	// Should be able to attempt again without locking immediately
	err = tracker.Track(ctx, userUID)
	if err != nil {
		t.Errorf("Track() after reset unexpected error = %v", err)
	}

	locked, _ = tracker.IsLocked(ctx, userUID)
	if locked {
		t.Error("IsLocked() = true after single attempt, want false")
	}
}

func TestRedisAttemptTracker_LockoutExpiry(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	threshold := 2
	lockoutDuration := 1 * time.Second // Redis truncates to minimum 1s
	counterTTL := 30 * time.Minute
	tracker := NewRedisAttemptTracker(client, threshold, lockoutDuration, counterTTL)
	userUID := "user-123"

	// Track attempts until lockout
	for i := 0; i < threshold; i++ {
		tracker.Track(ctx, userUID)
	}

	// Should be locked
	locked, _ := tracker.IsLocked(ctx, userUID)
	if !locked {
		t.Fatal("IsLocked() = false, want true (after threshold attempts)")
	}

	// Wait for lockout to expire (need > 1s due to Redis truncation)
	time.Sleep(1100 * time.Millisecond)

	// Should no longer be locked
	locked, _ = tracker.IsLocked(ctx, userUID)
	if locked {
		t.Error("IsLocked() = true, want false (after lockout expiry)")
	}
}

func TestRedisAttemptTracker_CounterExpiry(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 1 * time.Second // Redis truncates to minimum 1s
	tracker := NewRedisAttemptTracker(client, threshold, lockoutDuration, counterTTL)
	userUID := "user-123"

	// Make 2 attempts (below threshold)
	tracker.Track(ctx, userUID)
	tracker.Track(ctx, userUID)

	// Wait for counter to expire (need > 1s due to Redis truncation)
	time.Sleep(1100 * time.Millisecond)

	// Third attempt should not lock (counter was reset)
	err := tracker.Track(ctx, userUID)
	if err != nil {
		t.Errorf("Track() after counter expiry unexpected error = %v", err)
	}

	locked, _ := tracker.IsLocked(ctx, userUID)
	if locked {
		t.Error("IsLocked() = true, want false (counter expired)")
	}
}

func TestRedisAttemptTracker_MultipleUsers(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewRedisAttemptTracker(client, threshold, lockoutDuration, counterTTL)

	user1 := "user-1"
	user2 := "user-2"

	// Lock out user1
	for i := 0; i < threshold; i++ {
		tracker.Track(ctx, user1)
	}

	// user1 should be locked
	locked, _ := tracker.IsLocked(ctx, user1)
	if !locked {
		t.Error("IsLocked(user1) = false, want true")
	}

	// user2 should not be locked
	locked, _ = tracker.IsLocked(ctx, user2)
	if locked {
		t.Error("IsLocked(user2) = true, want false")
	}
}

func TestRedisAttemptTracker_Concurrent(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	threshold := 100
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewRedisAttemptTracker(client, threshold, lockoutDuration, counterTTL)
	userUID := "user-concurrent"

	// Launch concurrent attempts
	done := make(chan bool, threshold)
	for i := 0; i < threshold; i++ {
		go func() {
			tracker.Track(ctx, userUID)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < threshold; i++ {
		<-done
	}

	// Should be locked after all attempts
	locked, _ := tracker.IsLocked(ctx, userUID)
	if !locked {
		t.Error("IsLocked() = false, want true (after concurrent attempts)")
	}
}

func TestRedisAttemptTracker_GetFailedAttempts(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	threshold := 5
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewRedisAttemptTracker(client, threshold, lockoutDuration, counterTTL)
	userUID := "user-123"

	// No attempts initially
	count, err := tracker.(*RedisAttemptTracker).GetFailedAttempts(ctx, userUID)
	if err != nil {
		t.Fatalf("GetFailedAttempts() unexpected error = %v", err)
	}
	if count != 0 {
		t.Errorf("GetFailedAttempts() = %d, want 0", count)
	}

	// Make some attempts
	for i := 0; i < 3; i++ {
		tracker.Track(ctx, userUID)
	}

	count, err = tracker.(*RedisAttemptTracker).GetFailedAttempts(ctx, userUID)
	if err != nil {
		t.Fatalf("GetFailedAttempts() unexpected error = %v", err)
	}
	if count != 3 {
		t.Errorf("GetFailedAttempts() = %d, want 3", count)
	}
}

func TestRedisAttemptTracker_GetLockoutRemaining(t *testing.T) {
	client := redis.CreateTestRedisClient(t)
	ctx := context.Background()

	threshold := 2
	lockoutDuration := 2 * time.Second
	counterTTL := 30 * time.Minute
	tracker := NewRedisAttemptTracker(client, threshold, lockoutDuration, counterTTL)
	userUID := "user-123"

	// No lockout initially
	remaining, err := tracker.(*RedisAttemptTracker).GetLockoutRemaining(ctx, userUID)
	if err != nil {
		t.Fatalf("GetLockoutRemaining() unexpected error = %v", err)
	}
	if remaining != 0 {
		t.Errorf("GetLockoutRemaining() = %v, want 0", remaining)
	}

	// Lock the account
	for i := 0; i < threshold; i++ {
		tracker.Track(ctx, userUID)
	}

	// Check remaining time
	remaining, err = tracker.(*RedisAttemptTracker).GetLockoutRemaining(ctx, userUID)
	if err != nil {
		t.Fatalf("GetLockoutRemaining() unexpected error = %v", err)
	}
	if remaining < 1*time.Second || remaining > 2*time.Second {
		t.Errorf("GetLockoutRemaining() = %v, want ~2s", remaining)
	}
}
