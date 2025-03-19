package datamodel

import (
	"fmt"
	"slices"
)

type RequestValidator interface {
    Validate() error
}

type UidGetter interface {
    GetUID() string
    RequestValidator
}

func missingRequestBodyField(field string) error {
    return fmt.Errorf("Invalid request body: missing '%s' field", field)
}

var InvalidUID = missingRequestBodyField("uid")
var InvalidLogin = missingRequestBodyField("login")
var InvalidPassword = missingRequestBodyField("password")
var InvalidRoles = missingRequestBodyField("roles")

type AuthRequestBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (b *AuthRequestBody) Validate() error {
    if b.Login == "" {
        return InvalidLogin
    }

    if b.Password == "" {
        return InvalidPassword
    }

    return nil
}

type UidBody struct {
	UID string `json:"uid"`
}

func (b *UidBody) Validate() error {
    if b.UID == "" {
        return InvalidUID
    }

    return nil
}

func (body *UidBody) GetUID() string {
    return body.UID
}

type LoginBody struct {
	Login string `json:"login"`
}

func (b *LoginBody) Validate() error {
    if b.Login == "" {
        return InvalidLogin
    }

    return nil
}

type UidAndLoginBody struct {
	UID   string `json:"uid"`
	Login string `json:"login"`
}

func (b *UidAndLoginBody) Validate() error {
    if b.UID == "" {
        return InvalidUID
    }

    if b.Login == "" {
        return InvalidLogin
    }

    return nil
}

func (body *UidAndLoginBody) GetUID() string {
    return body.UID
}

type UidAndPasswordBody struct {
	UID      string `json:"uid"`
	Password string `json:"password"`
}

func (b *UidAndPasswordBody) Validate() error {
    if b.UID == "" {
        return InvalidUID
    }

    if b.Password == "" {
        return InvalidPassword
    }

    return nil
}

func (body *UidAndPasswordBody) GetUID() string {
    return body.UID
}

type UidAndRolesBody struct {
	UID   string   `json:"uid"`
	Roles []string `json:"roles"`
}

func (body *UidAndRolesBody) GetUID() string {
    return body.UID
}

func (b *UidAndRolesBody) Validate() error {
    if b.UID == "" {
        return InvalidUID
    }

    if len(b.Roles) == 0 {
        return InvalidRoles
    }

    if slices.Contains(b.Roles, "") {
        return InvalidRoles
    }

    return nil
}

