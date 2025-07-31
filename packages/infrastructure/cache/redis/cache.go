package redis

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var log = logger.NewSource("CACHE", logger.Default)

type driver struct {
    client *redis.Client
    isConnected bool
}

func New() *driver {
    return new(driver)
}

func (d *driver) Connect() {
    if d.isConnected {
        log.Panic("DB connection failed", "Connection already established", nil)
    }

	log.Info("Connecting to DB...", nil)

	d.client = redis.NewClient(&redis.Options{
		Addr:        config.Secret.CacheURI,
		Password:    config.Secret.CachePassword,
		DB:          config.Secret.CacheDB,
		ReadTimeout: config.Cache.SocketTimeout(),
    })

    ctx, cancel := defaultTimeoutContext()
    defer cancel()

    if err := d.client.Ping(ctx).Err(); err != nil {
        log.Panic("DB connection failed", err.Error(), nil)
    }

	log.Info("Connecting to DB: OK", nil)

    d.isConnected = true
}

func (d *driver) Close() *Error.Status {
    if !d.isConnected {
        return Error.NewStatusError(
            "connection not established",
            http.StatusInternalServerError,
        )
    }

    log.Info("Disconnecting from DB...", nil)

    if err := d.client.Close(); err != nil {
        return Error.NewStatusError(
            err.Error(),
            http.StatusInternalServerError,
        )
    }

    log.Info("Disconnecting from DB: OK", nil)

    d.isConnected = false

    return nil
}

func defaultTimeoutContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), config.Cache.OperationTimeout())
}

// x2 of defaultTimeoutContext()
func longTimeoutContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), config.Cache.OperationTimeout() * 2)
}

// Performs logging.
// Returns err converted to *Error.Status if err is not nil.
// If err is nil then just logs given action (with trace level).
func handleError(action string, err error) *Error.Status {
	if err != nil {
        if err == context.DeadlineExceeded {
            log.Error(
                "Request failed",
                "TIMEOUT: " + action,
				nil,
            )
        } else {
            log.Error(
                "Request failed",
                "Failed to "+action+": "+err.Error(),
				nil,
            )
        }
        return Error.StatusInternalError
	}

    log.Trace(action, nil)

    return nil
}

func (d *driver) Get(key string) (string, bool) {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	cachedData, err := d.client.Get(ctx, key).Result()
    if err == redis.Nil {
        log.Trace("Miss: " + key, nil)
        return "", false
    }

    return cachedData, handleError("Get: " + key, err) == nil
}

// IMPORTANT:
// go-redis driver can handle only this types:
// string, bool, []byte, int, int64, float64, time.Time
func(d *driver) Set(key string, value any) *Error.Status {
    // Alas, generics can't be used in methods
    // (it can be passed to a struct, but thats kinda strange and
    //  even so i failed to make it works as i want, so using type switch instead)
	switch v := value.(type) {
    case string, bool, []byte, int, int64, float64, time.Time:
        // Type allowed, do nothing and just go forward
	case uint32:
		if uint64(v) > uint64(math.MaxInt64) {
			return handleError("Set: ", fmt.Errorf("value overflows int64: %v", value))
		}
		value = int64(v)
	case uint64:
		if v > uint64(math.MaxInt64) {
			return handleError("Set: ", fmt.Errorf("value overflows int64: %v", value))
		}
		value = int64(v)
    default:
        return handleError("Set: ", fmt.Errorf("invalid cache value type: %T", value))
    }

    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := d.client.Set(ctx, key, value, config.Cache.TTL()).Err()

   return handleError("Set: " + key, err)
}

func (d *driver) Delete(keys ...string) *Error.Status {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := d.client.Unlink(ctx, keys...).Err()

    return handleError("Delete: " + strings.Join(keys, ","), err)
}

func (d *driver) FlushAll() *Error.Status {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

	err := d.client.FlushAll(ctx).Err()

    return handleError("Flush All", err)
}

const deletePatternAction = "Delete Pattern: "

func (d *driver) DeletePattern(pattern string) *Error.Status {
    // Initialize cursor for iteration
    var cursor uint64
    var keys []string
    var err error

    ctx, cancel := longTimeoutContext()
    defer cancel()

    // Use SCAN to find all keys matching the pattern
    for {
        if err := ctx.Err(); err != nil {
            return handleError(deletePatternAction, err)
        }

        keys, cursor, err = d.client.Scan(ctx, cursor, pattern, 100).Result()
        if err != nil {
			return handleError(
				deletePatternAction,
				errors.New("Error scanning keys: " + err.Error()),
			)
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
                    return handleError(deletePatternAction, ctxErr)
                }
				return handleError(
					deletePatternAction,
					errors.New("Error deleting keys: " + err.Error()),
				)
            }

            deleted := strconv.FormatInt(int64(len(keys)), 64)

            log.Trace("Deleted "+deleted+" keys with pattern: "+pattern, nil)
        }

        // Exit when cursor is 0 (no more keys to scan)
        if cursor == 0 {
            return handleError(deletePatternAction, nil)
        }
    }
}

const progressiveDeletePatternAction = "Progressive Delete Pattern"

func (d *driver) ProgressiveDeletePattern(pattern string) *Error.Status {
    const scanBatchSize = 500
    const unlinkBatchSize = 1000
    var cursor uint64

    batch := make([]string, 0, unlinkBatchSize)
    keysDeleted := 0

    scanCtx, cancelScanCtx := longTimeoutContext()
    defer cancelScanCtx()

    for {
        if err := scanCtx.Err(); err != nil {
            return handleError(progressiveDeletePatternAction, err)
        }

        keys, nextCursor, err := d.client.Scan(scanCtx, cursor, pattern, scanBatchSize).Result()
        if err != nil {
            if ctxErr := scanCtx.Err(); ctxErr != nil {
                return handleError(progressiveDeletePatternAction, err)
            }
			return handleError(
				progressiveDeletePatternAction,
				errors.New("Redis scan failed: " + err.Error()),
			)
        }

        batch = append(batch, keys...)

        if len(batch) > 0 && (len(batch) >= unlinkBatchSize || nextCursor == 0) {
            timeout := time.Duration(max(1, len(batch)/100)) * time.Millisecond

            ctx, cancel := context.WithTimeout(context.Background(), timeout)
            defer cancel()

            if _, err := d.client.Unlink(ctx, batch...).Result(); err != nil {
                if ctxErr := ctx.Err(); ctxErr != nil {
                    return handleError(progressiveDeletePatternAction, err)
                }
				return handleError(
					progressiveDeletePatternAction,
					errors.New("Batch unlink failed: " + err.Error()),
				)
            }

            keysDeleted += len(batch)
            batch = batch[:0] // reset batch
        }

        if nextCursor == 0 {
            deleted := strconv.FormatInt(int64(keysDeleted), 64)
            log.Trace("Deleted "+deleted+" keys matching "+pattern, nil)
            break
        }

        cursor = nextCursor
    }

    return nil
}

const progressiveDeleteKeysAction = "Progressive Delete"

func (d *driver) ProgressiveDelete(keys []string) *Error.Status {
	const batchSize = 1000

	for i := 0; i < len(keys); i += batchSize {
		end := min(len(keys), i + batchSize)

		batch := keys[i:end]
		ctx, cancel := longTimeoutContext()
		defer cancel()

		if _, err := d.client.Unlink(ctx, batch...).Result(); err != nil {
			return handleError(
				progressiveDeleteKeysAction,
				fmt.Errorf("batch %d-%d failed - %s", i, end, err.Error()),
			)
		}
	}

	return nil
}

