package cache

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

type RedisCache struct{ client *redis.Client }

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}
