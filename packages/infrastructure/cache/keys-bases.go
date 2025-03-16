package cache

var AnyUserById = "any_user_by_id"
var UserById = "user_by_id"
var DeletedUserById = "deleted_user_by_id"
var UserRolesById = "user_roles_by_id"
var UserByLogin = "user_by_login"
var AnyUserByLogin = "any_user_by_login"

var KeyBase = map[string]string {
    AnyUserById: AnyUserKeyPrefix + "id:",
    UserById: UserKeyPrefix + "id:",
    DeletedUserById: DeletedUserKeyPrefix + "id:",
    UserRolesById: UserKeyPrefix + "roles:",
    UserByLogin: UserKeyPrefix + "login:",
    AnyUserByLogin: AnyUserKeyPrefix + "login:",
}

