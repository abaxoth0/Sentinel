package user

import (
	"net/http"
	"regexp"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	"sentinel/packages/core"
	"strings"
)

type Property core.EntityProperty

const (
    IdProperty Property = "id"
    LoginProperty 		= "login"
    RolesProperty 		= "roles"
    PasswordProperty 	= "password"
    DeletedAtProperty 	= "deleted_at"
	VersionProperty 	= "version"
)

// Represents user deletion state, might be:
// deleted (deletedState), not deleted (notDeletedState), any (anyState)
type State byte

const (
    NotDeletedState State = iota
    DeletedState
    AnyState
)

var stateMap = map[State]string{
	NotDeletedState: "not deleted",
	DeletedState:	 "deleted",
	AnyState:		 "any",
}

func (s State) String() string {
	return stateMap[s]
}

const allowedSymbolsMsg = "Разрешённые символы: латинксие буквы, цифры от 0 до 9, спецсимволы '_', '-', '.', '@', '$', '!', '#'"

var ErrInvalidPasswordLength = Error.NewStatusError(
    "Пароль должен находится в диапозоне от 8 до 64 символов.",
    http.StatusBadRequest,
)
var ErrPasswordsContainsUnacceptableSymbols = Error.NewStatusError(
    "Пароль содержит недопустимые символы. " + allowedSymbolsMsg,
    http.StatusBadRequest,
    )
var ErrInvalidLoginLength = Error.NewStatusError(
    "Логин должен находиться в диапозоне от 5 до 72 символов.",
    http.StatusBadRequest,
)
var ErrInvalidEmailFormat = Error.NewStatusError(
    "Неверный логин: недопустимый формат E-Mail'а",
    http.StatusBadRequest,
)

var allowedSymbolsRegexp = regexp.MustCompile(`^[a-zA-Z0-9_\-\.@$!#]+$`)

func ValidatePassword(password string) *Error.Status {
	passwordSize := len(strings.ReplaceAll(password, " ", ""))

	// bcrypt can handle password with maximum size of 72 bytes
	if passwordSize < 8 || passwordSize > 64 {
		return ErrInvalidPasswordLength
    }

    if !allowedSymbolsRegexp.MatchString(password) {
        return ErrPasswordsContainsUnacceptableSymbols
    }

	return nil
}

func ValidateLogin(login string) *Error.Status {
    length := len(strings.ReplaceAll(login, " ", ""))

    if length < 5 || length > 72 {
        return ErrInvalidLoginLength
    }

	if err := validation.Email(login); err != nil {
		// If err is not nil then it maybe only Error.InvalidValie,
		// cuz login was already checked for zero or whitespaces value
		return ErrInvalidEmailFormat
	}

    return nil
}

