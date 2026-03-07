package resolver

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newMockRedis creates a new miniredis server and returns the redis client and cleanup function
func newMockRedis() (*redis.Client, func(), error) {
	s, err := miniredis.Run()
	if err != nil {
		return nil, nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cleanup := func() {
		client.Close()
		s.Close()
	}

	return client, cleanup, nil
}
