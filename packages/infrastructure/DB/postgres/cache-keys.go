package postgres

import "sentinel/packages/infrastructure/cache"

func userCacheKey(uid string, deleted bool) string {
    if deleted {
        return cache.DeletedUserKeyPrefix + "id:" + uid
    }
    return cache.UserKeyPrefix + "id:" + uid
}

func userRolesCacheKey(uid string) string {
    return cache.UserKeyPrefix + "roles:" + uid
}

