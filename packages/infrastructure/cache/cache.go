package cache

import (
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
    EncodeAndSet(key string, value interface{}) error
    Delete(keys ...string) error
    Drop() error
}

var Client client = redis.New()

