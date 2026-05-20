package sessioncache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Get(key string) (string, bool) {
	val, err := c.client.Get(context.Background(), key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

func (c *RedisCache) Set(key, value string, ttl time.Duration) {
	c.client.Set(context.Background(), key, value, ttl)
}
