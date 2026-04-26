package handler

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type rateLimiter struct {
	store rateLimitStore
}

func newRateLimiter(store rateLimitStore) *rateLimiter {
	return &rateLimiter{
		store: store,
	}
}

func (r *rateLimiter) AllowIP(ctx context.Context, ip string, limit int) bool {
	key := fmt.Sprintf("rate_limit:ip:%s", ip)

	return r.allow(ctx, key, limit, time.Minute)
}

func (r *rateLimiter) AllowEmail(ctx context.Context, email string, limit int) bool {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	key := fmt.Sprintf("rate_limit:email:%s", normalizedEmail)

	return r.allow(ctx, key, limit, 5*time.Minute)
}

func (r *rateLimiter) allow(ctx context.Context, key string, limit int, ttl time.Duration) bool {
	if limit <= 0 {
		return true
	}

	count, err := r.store.Incr(ctx, key, ttl)
	if err != nil {
		return true
	}

	return count <= int64(limit)
}
