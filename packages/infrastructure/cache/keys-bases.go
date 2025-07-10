package cache

const (
	AnyUserById = "any_user_by_id"
	UserById = "user_by_id"
	DeletedUserById = "deleted_user_by_id"
	UserRolesById = "user_roles_by_id"
	UserByLogin = "user_by_login"
	AnyUserByLogin = "any_user_by_login"
	UserBySessionID = "user_by_session_id"

	UserVersionByID = "user_version_by_id"

	SessionByID = "session_by_id"
)

var KeyBase = map[string]string {
    AnyUserById: AnyUserKeyPrefix + "id:",
    UserById: UserKeyPrefix + "id:",
    DeletedUserById: DeletedUserKeyPrefix + "id:",
    UserRolesById: UserKeyPrefix + "roles:",
    UserByLogin: UserKeyPrefix + "login:",
    AnyUserByLogin: AnyUserKeyPrefix + "login:",
	UserBySessionID: SessionKeyPrefix + "session:",

	UserVersionByID: UserKeyPrefix + "version:",

	SessionByID: SessionKeyPrefix + "id:",
}

