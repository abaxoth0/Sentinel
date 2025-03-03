package cache

import (
	"context"
	"fmt"
	"log"
	"sentinel/packages/infrastructure/config"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const UserKeyPrefix string = "user_"
const DeletedUserKeyPrefix string = "sd_user_"

var client *redis.Client
var isInit bool = false

func Init() {
    if isInit {
        panic("cache already initialized")
    }

	log.Println("[ CACHE ] Initializng...")

	client = redis.NewClient(&redis.Options{
		Addr:        config.Cache.URI,
		Password:    config.Cache.Password,
		DB:          config.Cache.DB,
		ReadTimeout: config.Cache.SocketTimeout,
    })

    ctx, cancel := defaultTimeoutContext()
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        panic(fmt.Sprintf("[ ERROR ] Failed to connect to Redis:\n%v\n", err))
    }

	log.Println("[ CACHE ] Initializng: OK")

    isInit = true
}

func defaultTimeoutContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), config.Cache.OperationTimeout)
}

func logAction(action string, err error) {
	if err != nil {
        if err == context.DeadlineExceeded {
            log.Println("[ CACHE ] TIMEOUT: " + action)
        } else {
            log.Printf("[ CACHE ] ERROR: Failed to '%s':\n%v\n", action, err)
        }
	} else {
		log.Println("[ CACHE ] " + action)
    }
}

func Get(key string) (string, bool) {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	cachedData, err := client.Get(ctx, key).Result()

    if err == redis.Nil {
        log.Println("[ CACHE ] Miss: " + key)
        return "", false
    }

    logAction("Get: " + key, err)

    return cachedData, err == nil
}

func Set[T string | bool | []byte | int | int64 | float64 | time.Time](key string, value T) error {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := client.Set(ctx, key, value, config.Cache.TTL).Err()

    logAction("Set: " + key, err)

	return err
}

func Delete(keys ...string) error {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := client.Del(ctx, keys...).Err()

    logAction("Delete: " + strings.Join(keys, ","), err)

	return err
}

func Drop() error {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := client.FlushAll(ctx).Err()

    logAction("Drop", err)

	return err
}

