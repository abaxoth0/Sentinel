package role

import externalerror "sentinel/packages/error"

// This is banned user
const RestrictedUser string = "restricted_user"

// This is an user who didn't yet activated his account
const UnconfirmedUser string = "unconfirmed_user"

const DefaultUser string = "user"

// Optional role
const Manager string = "manager"

const Support string = "support"

const Moderator string = "moderator"

// Can all
const Administrator string = "admin"

// Array with all roles
var List = [7]string{
	RestrictedUser,
	UnconfirmedUser,
	DefaultUser,
	Manager,
	Support,
	Moderator,
	Administrator,
}

func Verify(role string) *externalerror.Error {
	for _, r := range List {
		if r == role {
			return nil
		}
	}

	return externalerror.New("Ошибка авторизации: неверная роль, попробуйте переавторизоваться", 400)
}
