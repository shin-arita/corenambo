package handler

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type rateLimitStore interface {
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type redisRateLimitStore struct {
	client *redis.Client
}

func newRedisRateLimitStore(addr string) *redisRateLimitStore {
	return &redisRateLimitStore{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (s *redisRateLimitStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := s.client.TxPipeline()

	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}
