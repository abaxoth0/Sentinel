package redis

import (
	"context"
	"fmt"
	"log"
	"sentinel/packages/infrastructure/config"
	"sentinel/packages/presentation/data/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type driver struct {
    client *redis.Client
    isInit bool
}

func New() *driver {
    return new(driver)
}

func (d *driver) Init() {
    if d.isInit {
        panic("cache already initialized")
    }

	log.Println("[ CACHE ] Initializng...")

	d.client = redis.NewClient(&redis.Options{
		Addr:        config.Cache.URI,
		Password:    config.Cache.Password,
		DB:          config.Cache.DB,
		ReadTimeout: config.Cache.SocketTimeout,
    })

    ctx, cancel := defaultTimeoutContext()
    defer cancel()

    if err := d.client.Ping(ctx).Err(); err != nil {
        panic(fmt.Sprintf("[ ERROR ] Failed to connect to Redis:\n%v\n", err))
    }

	log.Println("[ CACHE ] Initializng: OK")

    d.isInit = true
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

func (d *driver) Get(key string) (string, bool) {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	cachedData, err := d.client.Get(ctx, key).Result()

    if err == redis.Nil {
        log.Println("[ CACHE ] Miss: " + key)
        return "", false
    }

    logAction("Get: " + key, err)

    return cachedData, err == nil
}

// go-redis driver can handle only this types:
// string, bool, []byte, int, int64, float64, time.Time
//
// use EncodeAndSet in case if value doesn't belong to any of this types
// (like structs, hashmaps, slices etc)
func(d *driver) Set(key string, value any) error {
    // Alas, generics can't be used in methods
    // (it can be passed to a struct, but thats kinda strange and
    //  even so i failed to make it works as i want, so using type switch instead)
    switch value.(type) {
    case string, bool, []byte, int, int64, float64, time.Time:
        // Type allowed, do nothing and just go forward
    default:
        return fmt.Errorf("invalid cache value type: %T", value)
    }

    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := d.client.Set(ctx, key, value, config.Cache.TTL).Err()

    logAction("Set: " + key, err)

	return err
}

func (d *driver) EncodeAndSet(key string, value interface{}) error {
    encodedData, err := json.Encode(value)

    if err != nil {
        return err
    }

    if err := d.Set(key, encodedData); err != nil {
        return err
    }

    return nil
}

func (d *driver) Delete(keys ...string) error {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := d.client.Del(ctx, keys...).Err()

    logAction("Delete: " + strings.Join(keys, ","), err)

	return err
}

func (d *driver) Drop() error {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := d.client.FlushAll(ctx).Err()

    logAction("Drop", err)

	return err
}

