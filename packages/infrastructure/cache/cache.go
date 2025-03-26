package cache

import (
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/cache/redis"
)

// Cache key must consist of 3 parts:
// prefix, like: user_
// origin, like: id:
// suffux, like: actual user id
//
// prefix + orgin called cache key base, like: user_id:
//
// For example, full cache key must look like that: user_id:2384

const UserKeyPrefix string = "user_"
const DeletedUserKeyPrefix string = "sd_user_"
const AnyUserKeyPrefix string = "any_user_"

type client interface {
    Init()
    Get(key string) (string, bool)
    Set(key string, value any) error
    EncodeAndSet(key string, value any) error
    Delete(keys ...string) error
    Drop() error
    // If 'err' is not nil, then deletes cache for each of specified 'keys'.
    // returns 'err'.
    DeleteOnError(err *Error.Status, keys ...string) *Error.Status
}

var Client client = redis.New()

