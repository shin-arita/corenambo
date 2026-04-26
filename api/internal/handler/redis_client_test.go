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
	// 存在しないポートに接続して強制エラー
	store := newRedisRateLimitStore("localhost:0")

	_, err := store.Incr(context.Background(), "test", time.Second)

	if err == nil {
		t.Fatal("expected error")
	}
}
