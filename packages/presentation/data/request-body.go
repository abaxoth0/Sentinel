package datamodel

import (
	"fmt"
	"sentinel/packages/common/validation"
	"strings"
)

func missingRequestBodyFieldValue(field string) error {
    return fmt.Errorf("Invalid request body: field '%s' has no value", field)
}

func invalidRequestBodyFieldValue(field string) error {
    return fmt.Errorf("Invalid request body: field '%s' has invalid value", field)
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
    Validate() error
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

func (b *UidBody) Validate() error {
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

func (b *PasswordBody) Validate() error {
    if b.Password == "" {
        return MissingPassword
    }
    if strings.ReplaceAll(b.Password, " ", "") == "" {
        return InvalidPassword
    }
    return nil
}

type LoginBody struct {
	Login string `json:"login"`
}

func (b *LoginBody) Validate() error {
    if b.Login == "" {
        return MissingLogin
    }
    if strings.ReplaceAll(b.Login, " ", "") == "" {
        return InvalidLogin
    }
    return nil
}

type RolesBody struct {
    Roles []string `json:"roles"`
}

func (b *RolesBody) Validate() error {
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

func (b *LoginPasswordBody) Validate() error {
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

func (b *UidLoginBody) Validate() error {
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

func (b *ChangePasswordBody) Validate() error {
    if err := b.PasswordBody.Validate(); err != nil {
        return err
    }
    if b.NewPassword == "" {
        return MissingNewPassword
    }
    if strings.ReplaceAll(b.NewPassword, " ", "") == "" {
        return InvalidNewPassword
    }
    return nil
}

type ChangeLoginBody struct {
    LoginBody `json:",inline"`
    PasswordBody `json:",inline"`
}

func (b *ChangeLoginBody) Validate() error {
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

func (b *ChangeRolesBody) Validate() error {
    if err := b.RolesBody.Validate(); err != nil {
        return err
    }
    if err := b.PasswordBody.Validate(); err != nil {
        return err
    }
    return nil
}

