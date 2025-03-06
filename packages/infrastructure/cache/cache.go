package cache

import (
	"sentinel/packages/infrastructure/cache/redis"
)

const UserKeyPrefix string = "user_"
const DeletedUserKeyPrefix string = "sd_user_"

type client interface {
    Init()
    Get(key string) (string, bool)
    Set(key string, value any) error
    EncodeAndSet(key string, value interface{}) error
    Delete(keys ...string) error
    Drop() error
}

var Client client = redis.New()

