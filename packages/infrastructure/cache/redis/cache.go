package redis

import (
	"context"
	"fmt"
	"log"
	"sentinel/packages/config"
	Error "sentinel/packages/errors"
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

// timeout is x5 of defaultTimeoutContext
func longTimeoutContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), config.Cache.OperationTimeout * 5)
}

func logged(action string, err error) error {
	if err != nil {
        if err == context.DeadlineExceeded {
            log.Println("[ CACHE ] TIMEOUT: " + action)
        } else {
            log.Printf("[ CACHE ] ERROR: Failed to '%s':\n%v\n", action, err)
        }
	} else {
		log.Println("[ CACHE ] " + action)
    }
    return err
}

func (d *driver) Get(key string) (string, bool) {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	cachedData, err := d.client.Get(ctx, key).Result()

    if err == redis.Nil {
        log.Println("[ CACHE ] Miss: " + key)
        return "", false
    }

    return cachedData, logged("Get: " + key, err) == nil
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
        return logged("Set: ", fmt.Errorf("invalid cache value type: %T", value))
    }

    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := d.client.Set(ctx, key, value, config.Cache.TTL).Err()

   return logged("Set: " + key, err)
}

func (d *driver) EncodeAndSet(key string, value any) error {
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

	err := d.client.Unlink(ctx, keys...).Err()

    return logged("Delete: " + strings.Join(keys, ","), err)
}

func (d *driver) FlushAll() error {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := d.client.FlushAll(ctx).Err()

    return logged("Flush All", err)
}

func (d *driver) DeleteOnError(err *Error.Status, keys ...string) *Error.Status {
    if err == nil {
        d.Delete(keys...)
    }

    return err
}

var deletePatternAction = "Delete Pattern: "

func (d *driver) DeletePattern(pattern string) error {
    // Initialize cursor for iteration
    var cursor uint64
    var keys []string
    var err error

    ctx, cancel := longTimeoutContext()
    defer cancel()

    // Use SCAN to find all keys matching the pattern
    for {
        if err := ctx.Err(); err != nil {
            return logged(deletePatternAction, err)
        }

        keys, cursor, err = d.client.Scan(ctx, cursor, pattern, 100).Result()

        if err != nil {
            return logged(deletePatternAction, fmt.Errorf("error scanning keys: %w", err))
        }

        // Delete all found keys in a pipeline for efficiency
        if len(keys) > 0 {
            pipeline := d.client.Pipeline()

            for _, key := range keys {
                pipeline.Unlink(ctx, key)
            }

            _, err = pipeline.Exec(ctx)

            if err != nil {
                if ctxErr := ctx.Err(); ctxErr != nil {
                    return logged(deletePatternAction, ctxErr)
                }
                return logged(deletePatternAction, fmt.Errorf("error deleting keys: %w", err))
            }

            log.Printf("[ CACHE ] Deleted %d keys with pattern: %s", len(keys), pattern)
        }

        // Exit when cursor is 0 (no more keys to scan)
        if cursor == 0 {
            return logged(deletePatternAction, nil)
        }
    }
}

var progressiveDeletePatternAction = "Progressive Delete Pattern"

func (d *driver) ProgressiveDeletePattern(pattern string) error {
    const scanBatchSize = 500
    const unlinkBatchSize = 1000
    var cursor uint64

    batch := make([]string, 0, unlinkBatchSize)
    keysDeleted := 0

    scanCtx, cancelScanCtx := longTimeoutContext()
    defer cancelScanCtx()

    for {
        if err := scanCtx.Err(); err != nil {
            return logged(progressiveDeletePatternAction, err)
        }

        keys, nextCursor, err := d.client.Scan(scanCtx, cursor, pattern, scanBatchSize).Result()

        if err != nil {
            if ctxErr := scanCtx.Err(); ctxErr != nil {
                return logged(progressiveDeletePatternAction, ctxErr)
            }
            return logged(progressiveDeletePatternAction, fmt.Errorf("redis scan failed: %w", err))
        }

        batch = append(batch, keys...)

        if len(batch) > 0 && (len(batch) >= unlinkBatchSize || nextCursor == 0) {
            timeout := time.Duration(max(1, len(batch)/100)) * time.Millisecond

            ctx, cancel := context.WithTimeout(context.Background(), timeout)
            defer cancel()

            if _, err := d.client.Unlink(ctx, batch...).Result(); err != nil {
                if ctxErr := ctx.Err(); ctxErr != nil {
                    return logged(progressiveDeletePatternAction, ctxErr)
                }
                return logged(progressiveDeletePatternAction, fmt.Errorf("batch unlink failed: %w", err))
            }

            keysDeleted += len(batch)
            batch = batch[:0] // reset batch
        }

        if nextCursor == 0 {
            log.Printf("[CACHE] Deleted %d keys matching '%s'", keysDeleted, pattern)
            break
        }

        cursor = nextCursor
    }

    return nil
}

