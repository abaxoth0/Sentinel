package token

import (
	"sentinel/packages/config"
	"sentinel/packages/util"
)

func generateAccessTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.AccessTokenTTL)
}

func generateRefreshTokenTtlTimestamp() int64 {
	return util.TimestampSinceNow(config.JWT.RefreshTokenTTL)
}
