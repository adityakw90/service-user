package security

import (
	"context"
	"testing"
	"time"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
)

func TestMemoryAttemptTracker_Track(t *testing.T) {
	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewMemoryAttemptTracker(threshold, lockoutDuration, counterTTL)
	defer func() {
		if mt, ok := tracker.(*MemoryAttemptTracker); ok {
			mt.Close()
		}
	}()

	ctx := context.Background()
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
			name:   "Second attempt does not lock",
			setup:  func() { tracker.Reset(ctx, userUID) },
			attempts: 2,
			wantLocked: false,
			wantErr:   false,
		},
		{
			name:   "Third attempt locks account",
			setup:  func() { tracker.Reset(ctx, userUID) },
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
		})
	}
}

func TestMemoryAttemptTracker_IsLocked(t *testing.T) {
	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewMemoryAttemptTracker(threshold, lockoutDuration, counterTTL)
	defer func() {
		if mt, ok := tracker.(*MemoryAttemptTracker); ok {
			mt.Close()
		}
	}()

	ctx := context.Background()
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

func TestMemoryAttemptTracker_Reset(t *testing.T) {
	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewMemoryAttemptTracker(threshold, lockoutDuration, counterTTL)
	defer func() {
		if mt, ok := tracker.(*MemoryAttemptTracker); ok {
			mt.Close()
		}
	}()

	ctx := context.Background()
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

func TestMemoryAttemptTracker_LockoutExpiry(t *testing.T) {
	threshold := 2
	lockoutDuration := 100 * time.Millisecond
	counterTTL := 30 * time.Minute
	tracker := NewMemoryAttemptTracker(threshold, lockoutDuration, counterTTL)
	defer func() {
		if mt, ok := tracker.(*MemoryAttemptTracker); ok {
			mt.Close()
		}
	}()

	ctx := context.Background()
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

	// Wait for lockout to expire
	time.Sleep(150 * time.Millisecond)

	// Should no longer be locked
	locked, _ = tracker.IsLocked(ctx, userUID)
	if locked {
		t.Error("IsLocked() = true, want false (after lockout expiry)")
	}
}

func TestMemoryAttemptTracker_CounterExpiry(t *testing.T) {
	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 100 * time.Millisecond
	tracker := NewMemoryAttemptTracker(threshold, lockoutDuration, counterTTL)
	defer func() {
		if mt, ok := tracker.(*MemoryAttemptTracker); ok {
			mt.Close()
		}
	}()

	ctx := context.Background()
	userUID := "user-123"

	// Make 2 attempts (below threshold)
	tracker.Track(ctx, userUID)
	tracker.Track(ctx, userUID)

	// Wait for counter to expire
	time.Sleep(150 * time.Millisecond)

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

func TestMemoryAttemptTracker_MultipleUsers(t *testing.T) {
	threshold := 3
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewMemoryAttemptTracker(threshold, lockoutDuration, counterTTL)
	defer func() {
		if mt, ok := tracker.(*MemoryAttemptTracker); ok {
			mt.Close()
		}
	}()

	ctx := context.Background()
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

func TestMemoryAttemptTracker_Concurrent(t *testing.T) {
	threshold := 100
	lockoutDuration := 5 * time.Minute
	counterTTL := 30 * time.Minute
	tracker := NewMemoryAttemptTracker(threshold, lockoutDuration, counterTTL)
	defer func() {
		if mt, ok := tracker.(*MemoryAttemptTracker); ok {
			mt.Close()
		}
	}()

	ctx := context.Background()
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

// Test interface compliance
func TestMemoryAttemptTracker_Interface(t *testing.T) {
	var _ portsec.AttemptTracker = NewMemoryAttemptTracker(3, 5*time.Minute, 30*time.Minute)
}
