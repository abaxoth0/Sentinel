package datamodel

import (
	"fmt"
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	"sentinel/packages/core/user"
	"strings"
)

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

var MissingUID = missingRequestBodyFieldValue("uid")
var InvalidUID = invalidRequestBodyFieldValue("uid")

var MissingLogin = missingRequestBodyFieldValue("login")
var InvalidLogin = invalidRequestBodyFieldValue("login")

var MissingPassword = missingRequestBodyFieldValue("password")
var InvalidPassword = invalidRequestBodyFieldValue("password")

var MissingNewPassword = missingRequestBodyFieldValue("newPassword")
var InvalidNewPassword = invalidRequestBodyFieldValue("newPassword")

var MissingRoles = missingRequestBodyFieldValue("roles")
var InvalidRoles = invalidRequestBodyFieldValue("roles")

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

type UidBody struct {
	UID string `json:"uid"`
}

func (b *UidBody) Validate() *Error.Status {
    if b.UID == "" {
        return MissingUID
    }
    if err := validation.UUID(b.UID); err != nil {
        return InvalidUID
    }
    return nil
}

func (body *UidBody) GetUID() string {
    return body.UID
}

type PasswordBody struct {
    Password string
}

func (b *PasswordBody) GetPassword() string {
    return b.Password
}

func (b *PasswordBody) Validate() *Error.Status {
    if strings.ReplaceAll(b.Password, " ", "") == "" {
        return MissingPassword
    }
    if err := user.VerifyPassword(b.Password); err != nil {
        return InvalidPassword
    }
    return nil
}

type LoginBody struct {
	Login string `json:"login"`
}

func (b *LoginBody) Validate() *Error.Status {
    if err := user.VerifyLogin(b.Login); err != nil {
        return err
    }
    return nil
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

type UidLoginBody struct {
    UidBody `json:",inline"`
    LoginBody `json:",inline"`
}

func (b *UidLoginBody) Validate() *Error.Status {
    if err := b.UidBody.Validate(); err != nil {
        return err
    }
    if err := b.LoginBody.Validate(); err != nil {
        return err
    }
    return nil
}

type ChangePasswordBody struct {
    PasswordBody `json:",inline"`
    NewPassword string `json:"newPassword"`
}

func (b *ChangePasswordBody) Validate() *Error.Status {
    if strings.ReplaceAll(b.NewPassword, " ", "") == "" {
        return MissingNewPassword
    }
    if err := user.VerifyPassword(b.NewPassword); err != nil {
        return InvalidNewPassword
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

