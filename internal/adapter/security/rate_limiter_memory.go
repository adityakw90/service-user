package security

import (
	"context"
	"sort"
	"sync"
	"time"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
)

// rateLimitState stores the request timestamps for a key.
type rateLimitState struct {
	mu         sync.Mutex
	timestamps []time.Time
}

// MemoryRateLimiter implements RateLimiter using in-memory storage.
// Suitable for development and testing environments.
type MemoryRateLimiter struct {
	mu          sync.RWMutex
	keys        map[string]*rateLimitState
	limit       int
	windowSize  time.Duration
	keyPrefix   string // prefix for internal keys (e.g., "ip:")
}

// NewMemoryRateLimiter creates a new in-memory rate limiter with configured limit and window.
func NewMemoryRateLimiter(limit int, windowSize time.Duration) portsec.RateLimiter {
	return &MemoryRateLimiter{
		keys:       make(map[string]*rateLimitState),
		limit:      limit,
		windowSize: windowSize,
		keyPrefix:  "ip:", // default prefix for IP-based rate limiting
	}
}

// Acquire attempts to acquire a rate limit slot for the given device IP.
func (m *MemoryRateLimiter) Acquire(ctx context.Context, deviceIp string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.keyPrefix + deviceIp
	state, exists := m.keys[key]
	if !exists {
		state = &rateLimitState{
			timestamps: make([]time.Time, 0, m.limit),
		}
		m.keys[key] = state
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-m.windowSize)

	// Remove timestamps outside the window
	cleaned := m.filterTimestamps(state.timestamps, windowStart)
	state.timestamps = cleaned

	// Check if limit reached
	currentCount := len(state.timestamps)
	if currentCount >= m.limit {
		return false, nil
	}

	// Add current request
	state.timestamps = append(state.timestamps, now)

	return true, nil
}

// Reset resets the rate limit counter for a device IP.
func (m *MemoryRateLimiter) Reset(ctx context.Context, deviceIp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.keyPrefix + deviceIp
	delete(m.keys, key)
	return nil
}

// filterTimestamps removes timestamps outside the window.
// Assumes state.mu is already held.
func (m *MemoryRateLimiter) filterTimestamps(timestamps []time.Time, windowStart time.Time) []time.Time {
	// Since timestamps are appended in order, we can use binary search
	idx := sort.Search(len(timestamps), func(i int) bool {
		return timestamps[i].After(windowStart)
	})

	if idx == len(timestamps) {
		// All timestamps are outside the window
		return timestamps[:0]
	}

	return timestamps[idx:]
}

// Cleanup removes expired entries from memory.
// This should be called periodically to prevent memory leaks.
func (m *MemoryRateLimiter) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-m.windowSize)

	for key, state := range m.keys {
		state.mu.Lock()

		// Remove all timestamps older than cutoff
		idx := sort.Search(len(state.timestamps), func(i int) bool {
			return state.timestamps[i].After(cutoff)
		})

		if idx >= len(state.timestamps) {
			// All timestamps expired, delete the key
			delete(m.keys, key)
		} else {
			state.timestamps = state.timestamps[idx:]
		}

		state.mu.Unlock()
	}
}
