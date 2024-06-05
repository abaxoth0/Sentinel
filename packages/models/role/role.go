package role

import externalerror "sentinel/packages/error"

type Role string

// This is banned user
const RestrictedUser Role = "restricted_user"

// This is an user who didn't yet activated his account
const UnconfirmedUser Role = "unconfirmed_user"

const DefaultUser Role = "user"

// Optional role
const Manager Role = "manager"

const Support Role = "support"

const Moderator Role = "moderator"

// Can all
const Administrator Role = "admin"

// Array with all roles
var List = [7]Role{
	RestrictedUser,
	UnconfirmedUser,
	DefaultUser,
	Manager,
	Support,
	Moderator,
	Administrator,
}

func (role Role) Verify() *externalerror.Error {
	for _, r := range List {
		if r == role {
			return nil
		}
	}

	return externalerror.New("Ошибка авторизации: неверная роль, попробуйте переавторизоваться", 400)
}
