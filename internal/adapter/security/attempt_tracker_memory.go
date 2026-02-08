package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
)

// userAttemptState stores the state of failed attempts for a user.
type userAttemptState struct {
	mu          sync.Mutex
	attempts    int
	lockedUntil time.Time
	lastAttempt time.Time
}

// MemoryAttemptTracker implements AttemptTracker using in-memory storage.
// Suitable for development and testing environments.
type MemoryAttemptTracker struct {
	mu              sync.RWMutex
	users           map[string]*userAttemptState
	threshold       int
	lockoutDuration time.Duration
	counterTTL      time.Duration
	stopCleanup     chan struct{}
}

// NewMemoryAttemptTracker creates a new in-memory attempt tracker.
func NewMemoryAttemptTracker(threshold int, lockoutDuration, counterTTL time.Duration) portsec.AttemptTracker {
	tracker := &MemoryAttemptTracker{
		users:           make(map[string]*userAttemptState),
		threshold:       threshold,
		lockoutDuration: lockoutDuration,
		counterTTL:      counterTTL,
		stopCleanup:     make(chan struct{}),
	}

	// Start background cleanup goroutine
	go tracker.cleanupExpiredEntries()

	return tracker
}

// Track records a failed attempt for the user.
func (m *MemoryAttemptTracker) Track(ctx context.Context, userUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.users[userUID]
	if !exists {
		state = &userAttemptState{}
		m.users[userUID] = state
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()

	// Check if the counter has expired (TTL based on last attempt)
	if !state.lastAttempt.IsZero() && now.Sub(state.lastAttempt) > m.counterTTL {
		// Counter expired, reset
		state.attempts = 0
	}

	state.attempts++
	state.lastAttempt = now

	// Check if threshold reached
	if state.attempts >= m.threshold {
		state.lockedUntil = now.Add(m.lockoutDuration)
		return fmt.Errorf("account locked after %d failed attempts", state.attempts)
	}

	return nil
}

// IsLocked checks if a user account is currently locked.
func (m *MemoryAttemptTracker) IsLocked(ctx context.Context, userUID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.users[userUID]
	if !exists {
		return false, nil
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.lockedUntil.IsZero() {
		return false, nil
	}

	// Check if lockout has expired
	if time.Now().After(state.lockedUntil) {
		// Lockout expired, clear state
		state.attempts = 0
		state.lockedUntil = time.Time{}
		return false, nil
	}

	return true, nil
}

// Reset clears the failed attempt counter for the user.
func (m *MemoryAttemptTracker) Reset(ctx context.Context, userUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.users[userUID]
	if !exists {
		return nil
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	// Reset counter and lockout
	state.attempts = 0
	state.lockedUntil = time.Time{}
	state.lastAttempt = time.Time{}

	return nil
}

// cleanupExpiredEntries periodically removes expired user states.
func (m *MemoryAttemptTracker) cleanupExpiredEntries() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCleanup:
			return
		}
	}
}

// cleanup removes entries that have expired.
func (m *MemoryAttemptTracker) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for userUID, state := range m.users {
		state.mu.Lock()

		// Remove if both lockout and counter TTL have expired
		lockoutExpired := state.lockedUntil.IsZero() || now.After(state.lockedUntil)
		counterExpired := state.lastAttempt.IsZero() || now.Sub(state.lastAttempt) > m.counterTTL

		if lockoutExpired && counterExpired {
			delete(m.users, userUID)
		}

		state.mu.Unlock()
	}
}

// Close stops the cleanup goroutine.
func (m *MemoryAttemptTracker) Close() {
	close(m.stopCleanup)
}
