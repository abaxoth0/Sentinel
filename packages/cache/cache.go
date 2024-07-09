package cache

import (
	"context"
	"errors"
	"log"
	"sentinel/packages/config"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client
var ctx context.Context

func Init() {
	log.Println("[ CACHE ] Initializng...")

	client = redis.NewClient(&redis.Options{
		Addr:        config.Cache.URI,
		Password:    config.Cache.Password,
		DB:          config.Cache.DB,
		ReadTimeout: config.Cache.SocketTimeout,
	})

	ctx = context.Background()

	log.Println("[ CACHE ] Initializng: OK")
}

func Get(key string) (string, bool) {
	r, e := client.Get(ctx, key).Result()

	if e != nil {
		if errors.Is(e, redis.Nil) {
			if config.Debug.Enabled {
				log.Println("[ CACHE ] Miss: " + key)
			}

			return r, false
		}

		log.Println("[ CACHE ] Critical error")
		log.Fatalln(e)
	}

	if config.Debug.Enabled {
		log.Println("[ CACHE ] Hit: " + key)
	}

	return r, true
}

func Set(key string, value any) error {
	e := client.Set(ctx, key, value, config.Cache.TTL).Err()

	if e != nil && config.Debug.Enabled {
		log.Println("[ CACHE ] Set: " + key)
	}

	return e
}

func Delete(keys ...string) error {
	e := client.Del(ctx, keys...).Err()

	if e != nil && config.Debug.Enabled {
		for _, key := range keys {
			log.Println("[ CACHE ] Delete: " + key)
		}
	}

	return e
}

func Drop() error {
	e := client.FlushAll(ctx).Err()

	if e != nil && config.Debug.Enabled {
		log.Println("[ CACHE ] Drop all")
	}

	return e
}

const UserKeyPrefix string = "user_"
