package utils

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// 令牌桶 Lua 脚本
// KEYS[1]: 限流 key，如 "lim:user:123"
// ARGV[1]: 桶容量 capacity
// ARGV[2]: 每秒补充速率 refill_rate
// ARGV[3]: 当前时间戳（秒，带小数）
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local tokens = tonumber(redis.call('hget', key, 'tokens'))
local last = tonumber(redis.call('hget', key, 'last_refill'))

if tokens == nil then
    tokens = capacity
    last = now
end

-- 按时间比例计算应补充的令牌
local delta = math.max(0, now - last)
local filled = math.min(capacity, tokens + delta * rate)

if filled >= 1 then
    redis.call('hset', key, 'tokens', filled - 1, 'last_refill', now)
    redis.call('expire', key, math.ceil(capacity / rate) + 1)
    return 1  -- 允许
else
    redis.call('expire', key, math.ceil(capacity / rate) + 1)
    return 0  -- 拒绝
end
`)

type DistributedRateLimiter struct {
	Capacity   float64
	RefillRate float64
	redisCli   *redis.Client
}

func NewDistributedRateLimiter(redisCli *redis.Client, cap, refill float64) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		Capacity:   cap,
		RefillRate: refill,
		redisCli:   redisCli,
	}
}

func (l *DistributedRateLimiter) Close() error {
	return l.redisCli.Close()
}

func (l *DistributedRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := float64(time.Now().UnixNano()) / 1e9
	result, err := tokenBucketScript.Run(ctx, l.redisCli, []string{key},
		l.Capacity, l.RefillRate, now).Int()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}
