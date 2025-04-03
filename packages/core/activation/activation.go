package activation

import "sentinel/packages/core"

type Property core.EntityProperty

const (
    IdProperty Property = "id"
    UserLoginProperty Property = "user_login"
    TokenProperty Property = "token"
    ExpiresAtProperty Property = "expires_at"
    CreatedAtProperty Property = "created_at"
)

