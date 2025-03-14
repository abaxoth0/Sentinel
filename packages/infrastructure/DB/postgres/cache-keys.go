package postgres

import "sentinel/packages/infrastructure/cache"

var anyUserCacheKeyBase = cache.AnyUserKeyPrefix + "id:"
var userCacheKeyBase = cache.UserKeyPrefix + "id:"
var deletedUserCacheKeyBase = cache.DeletedUserKeyPrefix + "id:"
var userRolesCacheKeyBase = cache.UserKeyPrefix + "roles:"

