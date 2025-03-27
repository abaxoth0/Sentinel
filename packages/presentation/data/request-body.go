package datamodel

import (
	"fmt"
	"slices"
)

func requestBodyFieldMissingValue(field string) error {
    return fmt.Errorf("Invalid request body: field '%s' has no value", field)
}

func requestBodyFieldInvalidValue(field string) error {
    return fmt.Errorf("Invalid request body: field '%s' has invalid value", field)
}

var MissingUID = requestBodyFieldMissingValue("uid")
var MissingLogin = requestBodyFieldMissingValue("login")
var MissingPassword = requestBodyFieldMissingValue("password")
var MissingRoles = requestBodyFieldMissingValue("roles")
var InvalidRoles = requestBodyFieldInvalidValue("roles")

type RequestValidator interface {
    Validate() error
}

type UidGetter interface {
    GetUID() string
    RequestValidator
}

type UidBody struct {
	UID string `json:"uid"`
}

func (b *UidBody) Validate() error {
    if b.UID == "" {
        return MissingUID
    }
    return nil
}

func (body *UidBody) GetUID() string {
    return body.UID
}

type PasswordBody struct {
    Password string
}

func (b *PasswordBody) Validate() error {
    if b.Password == "" {
        return MissingUID
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
    return nil
}

type RolesBody struct {
    Roles []string `json:"roles"`
}

func (b *RolesBody) Validate() error {
    if len(b.Roles) == 0 {
        return MissingRoles
    }
    if slices.Contains(b.Roles, "") {
        return InvalidRoles
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

type UidPasswordBody struct {
    UidBody `json:",inline"`
    PasswordBody `json:",inline"`
}

func (b *UidPasswordBody) Validate() error {
    if err := b.UidBody.Validate(); err != nil {
        return err
    }
    if err := b.PasswordBody.Validate(); err != nil {
        return err
    }
    return nil
}

type UidRolesBody struct {
    UidBody `json:",inline"`
    RolesBody `json:",inline"`
}

func (b *UidRolesBody) Validate() error {
    if err := b.UidBody.Validate(); err != nil {
        return err
    }
    if err := b.RolesBody.Validate(); err != nil {
        return err
    }
    return nil
}

