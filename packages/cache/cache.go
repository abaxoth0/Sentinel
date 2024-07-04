package cache

import (
	"context"
	"sentinel/packages/config"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client
var ctx context.Context

func Init() {
	client = redis.NewClient(&redis.Options{
		Addr:        config.Cache.URI,
		Password:    config.Cache.Password,
		DB:          config.Cache.DB,
		ReadTimeout: config.Cache.SocketTimeout,
	})

	ctx = context.Background()
}

func Get(key string) (string, error) {
	return client.Get(ctx, key).Result()
}

func Set(key string, value any) error {
	return client.Set(ctx, key, value, config.Cache.TTL).Err()
}

func Delete(keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

func Drop() error {
	return client.FlushAll(ctx).Err()
}
