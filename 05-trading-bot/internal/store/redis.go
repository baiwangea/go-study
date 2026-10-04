package store

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAddrFromEnv 读取 BOT_REDIS_ADDR（形如 127.0.0.1:6379），空则返回 ""。
func RedisAddrFromEnv() string { return os.Getenv("BOT_REDIS_ADDR") }

// RedisDeduper 用 SetNX 做跨进程幂等：这是内存版解决不了的那一半（多实例部署）。
// key 带 TTL，避免无限堆积；TTL 要大于「可能重复投递的最长时间窗」。
type RedisDeduper struct {
	cli *redis.Client
	ttl time.Duration
}

// NewRedisDeduper 会 PING 一次，把「连不上」暴露在建对象时而不是第一次去重时。
func NewRedisDeduper(addr string, db int, ttl time.Duration) (*RedisDeduper, error) {
	cli := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := cli.Ping(ctx).Err(); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("Redis 不可用(%s): %w", addr, err)
	}
	return &RedisDeduper{cli: cli, ttl: ttl}, nil
}

// MarkOnce 满足 executor.Deduper：第一次返回 true，之后都是 false。
func (r *RedisDeduper) MarkOnce(ctx context.Context, key string) (bool, error) {
	ok, err := r.cli.SetNX(ctx, "bot:done:"+key, 1, r.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("SetNX 失败: %w", err)
	}
	return ok, nil
}

func (r *RedisDeduper) Name() string { return "Redis SetNX（跨进程有效）" }

func (r *RedisDeduper) Close() error { return r.cli.Close() }
