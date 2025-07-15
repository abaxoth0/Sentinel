package requestbody

import (
	"fmt"
	"net/http"
	Error "sentinel/packages/common/errors"
	"slices"
	"strings"
)

/*
   IMPORTANT
   All kind of validation done in methods inside of this module is
   related to transport layer, which means:
   1) Validation checks only if value persist and it's not empty, cuz
      all what transport layer should do - is just be intermediary between
      user and business logic.
   2) All other kind of validation must be done on business logic layer
      e.g. - check if password or login doesn't include some unacceptable symbols
      or if user ID has correct format.
*/

func missingFieldValue(field string) *Error.Status {
    return Error.NewStatusError(
        fmt.Sprintf("Invalid request body: field '%s' has no value", field),
        http.StatusBadRequest,
    )
}

func invalidFieldValue(field string) *Error.Status {
    return Error.NewStatusError(
        fmt.Sprintf("Invalid request body: field '%s' has invalid value", field),
        http.StatusBadRequest,
    )
}

var ErrorMissingLogin = missingFieldValue("login")
var ErrorInvalidLogin = invalidFieldValue("login")

var ErrorMissingPassword = missingFieldValue("password")
var ErrorInvalidPassword = invalidFieldValue("password")

var ErrorMissingNewPassword = missingFieldValue("newPassword")
var ErrorInvalidNewPassword = invalidFieldValue("newPassword")

var ErrorMissingRoles = missingFieldValue("roles")
var ErrorInvalidRoles = invalidFieldValue("roles")

var ErrorMissingUserIDs = missingFieldValue("IDs")
var ErrorInvalidUserIDs = invalidFieldValue("IDs")

var invalidField map[string]*Error.Status = map[string]*Error.Status{
    "login": ErrorInvalidLogin,
    "password": ErrorInvalidPassword,
    "newPassword": ErrorInvalidNewPassword,
    "roles": ErrorInvalidRoles,
}

var missingField map[string]*Error.Status = map[string]*Error.Status{
    "login": ErrorMissingLogin,
    "password": ErrorMissingPassword,
    "newPassword": ErrorMissingNewPassword,
    "roles": ErrorMissingRoles,
}

func validateStr(field string, value string) *Error.Status {
    if value == "" {
        return missingField[field]
    }
    if strings.ReplaceAll(value, " ", "") == ""{
        return invalidField[field]
    }
    return nil
}

type Validator interface {
    Validate() *Error.Status
}

type PasswordGetter interface {
    GetPassword() string
}

type UpdateUser interface {
    PasswordGetter
    Validator
}

// swagger:model UserPasswordRequest
type UserPassword struct {
	Password string `json:"password" example:"your-password"`
}

func (b *UserPassword) GetPassword() string {
    return b.Password
}

func (b *UserPassword) Validate() *Error.Status {
    return validateStr("password", b.Password)
}

// swagger:model UserLoginRequest
type UserLogin struct {
	Login string `json:"login" example:"admin@mail.com"`
}

func (b *UserLogin) Validate() *Error.Status {
    return validateStr("login", b.Login)
}

// swagger:model UserRolesRequest
type UserRoles struct {
	Roles []string `json:"roles" example:"user,moderator"`
}

func (b *UserRoles) Validate() *Error.Status {
    if len(b.Roles) == 0 {
        return ErrorMissingRoles
    }
    for _, role := range b.Roles {
        if strings.ReplaceAll(role, " ", "") == "" {
            return ErrorInvalidRoles
        }
    }
    return nil
}

// swagger:model UserLoginAndPasswordRequest
type LoginAndPassword struct {
    UserLogin 		`json:",inline"`
    UserPassword 	`json:",inline"`
}

func (b *LoginAndPassword) Validate() *Error.Status {
    if err := b.UserLogin.Validate(); err != nil {
        return err
    }
    if err := b.UserPassword.Validate(); err != nil {
        return err
    }
    return nil
}

// swagger:model UserChangePasswordRequest
type ChangePassword struct {
    UserPassword 		`json:",inline"`
	NewPassword string 	`json:"newPassword" example:"your-new-password"`
}

func (b *ChangePassword) Validate() *Error.Status {
    if err := validateStr("newPassword", b.NewPassword); err != nil {
        return err
    }
    if err := b.UserPassword.Validate(); err != nil {
        return err
    }
    return nil
}

// swagger:model UserChangeLoginRequest
type ChangeLogin struct {
    UserLogin 		`json:",inline"`
    UserPassword 	`json:",inline"`
}

func (b *ChangeLogin) Validate() *Error.Status {
    if err := b.UserLogin.Validate(); err != nil {
        return err
    }
    if err := b.UserPassword.Validate(); err != nil {
        return err
    }
    return nil
}

// swagger:model UserChangeRolesRequest
type ChangeRoles struct {
    UserRoles 		`json:",inline"`
    UserPassword 	`json:",inline"`
}

func (b *ChangeRoles) Validate() *Error.Status {
    if err := b.UserRoles.Validate(); err != nil {
        return err
    }
    if err := b.UserPassword.Validate(); err != nil {
        return err
    }
    return nil
}

// swagger:model UsersIDsRequest
type UsersIDs struct {
	IDs []string `json:"id" example:"cef85e5a-5a5f-42d0-81bd-1650391c0e82,9bc87af1-5f92-4d8c-bf41-7ade642c5a91"`
}

func (b *UsersIDs) Validate() *Error.Status {
	if b.IDs == nil || len(b.IDs) == 0 {
		return ErrorMissingUserIDs
	}
	if slices.Contains(b.IDs, "") {
		return ErrorInvalidUserIDs
	}
	return nil
}

