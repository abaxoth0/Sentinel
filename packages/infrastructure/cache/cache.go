package cache

import (
	Error "sentinel/packages/common/errors"
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

const (
	UserKeyPrefix 		 	= "user_"
	DeletedUserKeyPrefix 	= "sd_user_"
	AnyUserKeyPrefix 	 	= "any_user_"
	SessionKeyPrefix 	 	= "session_"
	RevokedSessionKeyPrefix = "revoked_session_"
	LocationKeyPrefix 		= "location_"
)

type client interface {
    Connect()
    Close() *Error.Status
    Get(key string) (string, bool)
    Set(key string, value any) *Error.Status
    Delete(keys ...string) *Error.Status
    FlushAll() *Error.Status
    // Deletes cache entries whose keys match the pattern.
    // When need to delete a lot of entries consider using ProgressiveDeletePattern.
    DeletePattern(pattern string) *Error.Status
    // Do the same as DeletePattern, but more optimized for deleting a large amount of entries
    ProgressiveDeletePattern(pattern string) *Error.Status
	// Same as Delete, but uses batch processing
	ProgressiveDelete(keys []string) *Error.Status
}

var Client client = redis.New()

