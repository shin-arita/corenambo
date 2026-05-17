package handler

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func testRedisURL() string {
	if u := os.Getenv("REDIS_URL"); u != "" {
		return u
	}
	return "redis://redis:6379/1"
}

func TestRedisRateLimitStoreIncr(t *testing.T) {
	store := newRedisRateLimitStore(testRedisURL())
	key := fmt.Sprintf("test:rate_limit:%d", time.Now().UnixNano())

	count, err := store.Incr(context.Background(), key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("unexpected count: %d", count)
	}

	count, err = store.Incr(context.Background(), key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatalf("unexpected count: %d", count)
	}
}

func TestRedisRateLimitStoreIncrError(t *testing.T) {
	store := newRedisRateLimitStore("redis://localhost:0/1")

	_, err := store.Incr(context.Background(), "test", time.Second)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewRedisRateLimitStorePanicOnInvalidURL(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid Redis URL")
		}
	}()
	newRedisRateLimitStore("not-a-valid-url")
}

func TestRedisRateLimitStoreTTLNotReset(t *testing.T) {
	store := newRedisRateLimitStore(testRedisURL())
	key := fmt.Sprintf("test:rate_limit:ttl:%d", time.Now().UnixNano())

	// 1回目: キーを作成しTTLを3秒にセット
	_, err := store.Incr(context.Background(), key, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// 1.5秒待機（TTL残り約1.5秒）
	time.Sleep(1500 * time.Millisecond)

	// 2回目: ExpireNX のためTTLはリセットされない
	_, err = store.Incr(context.Background(), key, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// 2秒追加待機: TTLがリセットされていなければキーは失効している
	time.Sleep(2 * time.Second)

	// キーが失効していれば新規作成でcount=1
	count, err := store.Incr(context.Background(), key, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("TTL was reset: expected count=1 after expiry but got %d", count)
	}
}
