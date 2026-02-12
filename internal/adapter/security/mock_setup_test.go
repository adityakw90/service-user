package security

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newMockRedis creates a new miniredis server and returns the redis client, miniredis instance, and cleanup function
func newMockRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()

	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create test redis connection: %v", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	t.Cleanup(func() {
		client.Close()
		s.Close()
	})

	return client, s
}
