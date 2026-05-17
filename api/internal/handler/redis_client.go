package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type rateLimitStore interface {
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type redisRateLimitStore struct {
	client *redis.Client
}

// newRedisRateLimitStore constructs a store from a Redis URL.
// Panics on an invalid URL: this is a startup configuration error, not a runtime error,
// so a panic here prevents the server from starting with a broken Redis config.
func newRedisRateLimitStore(redisURL string) *redisRateLimitStore {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(fmt.Sprintf("invalid Redis URL %q: %v", redisURL, err))
	}
	return &redisRateLimitStore{
		client: redis.NewClient(opt),
	}
}

func (s *redisRateLimitStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := s.client.TxPipeline()

	incr := pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}
