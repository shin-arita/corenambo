package handler

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRedisRateLimitStoreIncr(t *testing.T) {
	store := newRedisRateLimitStore("redis:6379")
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
	store := newRedisRateLimitStore("localhost:0")

	_, err := store.Incr(context.Background(), "test", time.Second)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRedisRateLimitStoreTTLNotReset(t *testing.T) {
	store := newRedisRateLimitStore("redis:6379")
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
