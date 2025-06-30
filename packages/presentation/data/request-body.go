package datamodel

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

func missingRequestBodyFieldValue(field string) *Error.Status {
    return Error.NewStatusError(
        fmt.Sprintf("Invalid request body: field '%s' has no value", field),
        http.StatusBadRequest,
    )
}

func invalidRequestBodyFieldValue(field string) *Error.Status {
    return Error.NewStatusError(
        fmt.Sprintf("Invalid request body: field '%s' has invalid value", field),
        http.StatusBadRequest,
    )
}

var MissingLogin = missingRequestBodyFieldValue("login")
var InvalidLogin = invalidRequestBodyFieldValue("login")

var MissingPassword = missingRequestBodyFieldValue("password")
var InvalidPassword = invalidRequestBodyFieldValue("password")

var MissingNewPassword = missingRequestBodyFieldValue("newPassword")
var InvalidNewPassword = invalidRequestBodyFieldValue("newPassword")

var MissingRoles = missingRequestBodyFieldValue("roles")
var InvalidRoles = invalidRequestBodyFieldValue("roles")

var MissingUserIDs = missingRequestBodyFieldValue("IDs")
var InvalidUserIDs = invalidRequestBodyFieldValue("IDs")

var invalidField map[string]*Error.Status = map[string]*Error.Status{
    "login": InvalidLogin,
    "password": InvalidPassword,
    "newPassword": InvalidNewPassword,
    "roles": InvalidRoles,
}

var missingField map[string]*Error.Status = map[string]*Error.Status{
    "login": MissingLogin,
    "password": MissingPassword,
    "newPassword": MissingNewPassword,
    "roles": MissingRoles,
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

type RequestValidator interface {
    Validate() *Error.Status
}

type PasswordGetter interface {
    GetPassword() string
}

type UpdateUserRequestBody interface {
    PasswordGetter
    RequestValidator
}

type PasswordBody struct {
    Password string
}

func (b *PasswordBody) GetPassword() string {
    return b.Password
}

func (b *PasswordBody) Validate() *Error.Status {
    return validateStr("password", b.Password)
}

type LoginBody struct {
	Login string `json:"login"`
}

func (b *LoginBody) Validate() *Error.Status {
    return validateStr("login", b.Login)
}

type RolesBody struct {
    Roles []string `json:"roles"`
}

func (b *RolesBody) Validate() *Error.Status {
    if len(b.Roles) == 0 {
        return MissingRoles
    }
    for _, role := range b.Roles {
        if strings.ReplaceAll(role, " ", "") == "" {
            return InvalidRoles
        }
    }
    return nil
}

type LoginPasswordBody struct {
    LoginBody `json:",inline"`
    PasswordBody `json:",inline"`
}

func (b *LoginPasswordBody) Validate() *Error.Status {
    if err := b.LoginBody.Validate(); err != nil {
        return err
    }
    if err := b.PasswordBody.Validate(); err != nil {
        return err
    }
    return nil
}

type ChangePasswordBody struct {
    PasswordBody `json:",inline"`
    NewPassword string `json:"newPassword"`
}

func (b *ChangePasswordBody) Validate() *Error.Status {
    if err := validateStr("newPassword", b.NewPassword); err != nil {
        return err
    }
    if err := b.PasswordBody.Validate(); err != nil {
        return err
    }
    return nil
}

type ChangeLoginBody struct {
    LoginBody `json:",inline"`
    PasswordBody `json:",inline"`
}

func (b *ChangeLoginBody) Validate() *Error.Status {
    if err := b.LoginBody.Validate(); err != nil {
        return err
    }
    if err := b.PasswordBody.Validate(); err != nil {
        return err
    }
    return nil
}

type ChangeRolesBody struct {
    RolesBody `json:",inline"`
    PasswordBody `json:",inline"`
}

func (b *ChangeRolesBody) Validate() *Error.Status {
    if err := b.RolesBody.Validate(); err != nil {
        return err
    }
    if err := b.PasswordBody.Validate(); err != nil {
        return err
    }
    return nil
}

type UserIDsBody struct {
	IDs []string `json:"id"`
}

func (b *UserIDsBody) Validate() *Error.Status {
	if b.IDs == nil || len(b.IDs) == 0 {
		return MissingUserIDs
	}
	if slices.Contains(b.IDs, "") {
		return InvalidUserIDs
	}
	return nil
}
