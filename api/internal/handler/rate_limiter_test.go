package handler

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRateLimitStore struct {
	count int64
	err   error
	key   string
	ttl   time.Duration
}

func (m *mockRateLimitStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	m.key = key
	m.ttl = ttl

	if m.err != nil {
		return 0, m.err
	}

	m.count++

	return m.count, nil
}

func TestRateLimiterAllowIP(t *testing.T) {
	store := &mockRateLimitStore{}
	limiter := newRateLimiter(store)

	if !limiter.AllowIP(context.Background(), "127.0.0.1", 2) {
		t.Fatal("first request should be allowed")
	}

	if !limiter.AllowIP(context.Background(), "127.0.0.1", 2) {
		t.Fatal("second request should be allowed")
	}

	if limiter.AllowIP(context.Background(), "127.0.0.1", 2) {
		t.Fatal("third request should be denied")
	}

	if store.key != "rate_limit:ip:127.0.0.1" {
		t.Fatalf("unexpected key: %s", store.key)
	}

	if store.ttl != time.Minute {
		t.Fatalf("unexpected ttl: %v", store.ttl)
	}
}

func TestRateLimiterAllowEmail(t *testing.T) {
	store := &mockRateLimitStore{}
	limiter := newRateLimiter(store)

	if !limiter.AllowEmail(context.Background(), " Test@Example.com ", 1) {
		t.Fatal("first request should be allowed")
	}

	if limiter.AllowEmail(context.Background(), " Test@Example.com ", 1) {
		t.Fatal("second request should be denied")
	}

	if store.key != "rate_limit:email:test@example.com" {
		t.Fatalf("unexpected key: %s", store.key)
	}

	if store.ttl != 5*time.Minute {
		t.Fatalf("unexpected ttl: %v", store.ttl)
	}
}

func TestRateLimiterAllowWhenLimitIsZero(t *testing.T) {
	store := &mockRateLimitStore{}
	limiter := newRateLimiter(store)

	if !limiter.AllowIP(context.Background(), "127.0.0.1", 0) {
		t.Fatal("request should be allowed")
	}

	if store.count != 0 {
		t.Fatalf("store should not be called: %d", store.count)
	}
}

func TestRateLimiterAllowWhenStoreError(t *testing.T) {
	store := &mockRateLimitStore{
		err: errors.New("redis error"),
	}
	limiter := newRateLimiter(store)

	if !limiter.AllowIP(context.Background(), "127.0.0.1", 1) {
		t.Fatal("request should be allowed on store error")
	}
}
