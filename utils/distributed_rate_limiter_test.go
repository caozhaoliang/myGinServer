package utils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestNewDistributed(t *testing.T) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	// 限流参数
	key := "lim:user:123"
	capacity := 100.0  // 桶容量
	refillRate := 10.0 // 每秒补充 10 个令牌
	l := NewDistributedRateLimiter(rdb, capacity, refillRate)

	// 模拟请求
	for i := 0; i < 15; i++ {
		allow, err := l.Allow(ctx, key)
		if err != nil {
			continue
		}
		if allow {
			fmt.Println("允许")
		} else {
			fmt.Println("禁止")
		}
		time.Sleep(50 * time.Millisecond)
	}

}
