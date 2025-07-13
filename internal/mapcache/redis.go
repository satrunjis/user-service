package mapcache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log/slog"
	"time"
)

type RedisCache struct {
	client *redis.Client
	logger *slog.Logger
	ttl    time.Duration // >= 7 days
}

func Init(ctx context.Context, url *string, logger *slog.Logger, ttl time.Duration) (*RedisCache, error) {
	const op = "redis.Init"
	log := logger.With("operation", op)
	log.Debug("initializing Redis client", "url", url, "ttl", ttl)

	cfg := &redis.Options{
		Addr:     *url,
		Password: "",
		DB:       1,
		PoolSize: 10,
	}

	client := redis.NewClient(cfg)

	start := time.Now()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Error("connection test failed", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	} else {
		log.Debug("client created", "duration", time.Since(start))
	}

	log.Info("redis initialized")

	if ttl < 7*24*time.Hour {
		log.WarnContext(ctx, "TTL is less than 7 days. The parameter has been forcibly changed")
		ttl = 7 * 24 * time.Hour
	}

	return &RedisCache{client: client, logger: logger, ttl: ttl}, nil
}
func (c *RedisCache) Get(ctx context.Context, key *string) (*[]byte, error) {
	result, err := c.client.GetEx(ctx, *key, c.ttl).Bytes()
	return &result, err
}
func (c *RedisCache) Set(ctx context.Context, key *string, data *[]byte) error {
	return c.client.Set(ctx, *key, *data, c.ttl).Err()
}
func (c *RedisCache) Close() error {
	return c.client.Close()
}
