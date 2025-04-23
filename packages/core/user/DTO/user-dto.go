package userdto

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	"slices"
	"time"
)

type Any interface {
    IsDeleted() bool
}

type Basic struct {
    ID           string    `json:"id"`
	Login        string    `json:"login"`
	Password     string    `json:"password"`
	Roles        []string  `json:"roles"`
	DeletedAt    time.Time `json:"deletedAt"`
}

func (dto *Basic) IsDeleted() bool {
    return !dto.DeletedAt.IsZero()
}

func (dto *Basic) IsActive() bool {
    return !slices.Contains(dto.Roles, "unconfirmed_user")
}

type Extended struct {
    ID           string    `json:"id"`
	Login        string    `json:"login"`
	Password     string    `json:"password"`
	Roles        []string  `json:"roles"`
	DeletedAt    time.Time `json:"deletedAt"`
    CreatedAt    time.Time `json:"createdAt"`
}

func (dto *Extended) IsDeleted() bool {
    return !dto.DeletedAt.IsZero()
}

func (dto *Extended) ToBasic() *Basic {
    return &Basic{
        ID: dto.ID,
        Login: dto.Login,
        Password: dto.Password,
        Roles: dto.Roles,
        DeletedAt: dto.DeletedAt,
    }
}

type Audit struct {
    ID               string    `json:"id"`
    ChangedUserID    string    `json:"changedUserID"`
    ChangedByUserID  string    `json:"changedByUserID"`
    Operation        string    `json:"operation"`
    Login            string    `json:"login"`
	Password         string    `json:"password"`
	Roles            []string  `json:"roles"`
	DeletedAt        time.Time `json:"deletedAt"`
    ChangedAt        time.Time `json:"changedAt"`
}

func (dto *Audit) IsDeleted() bool {
    return !dto.DeletedAt.IsZero()
}

type Payload struct {
	ID    string   `json:"id"`
	Login string   `json:"login"`
	Roles []string `json:"roles"`
}

type Filter struct {
	TargetUID      string
	RequesterUID   string
	RequesterRoles []string
}

func (f *Filter) ValidateTargetUID() *Error.Status {
    if err := validation.UUID(f.TargetUID); err != nil {
        return err.ToStatus(
            "Target user ID is not specified",
            "Invalid target user ID",
        )
    }
    return nil
}

func (f *Filter) ValidateRequesterUID() *Error.Status {
    if err := validation.UUID(f.RequesterUID); err != nil {
        return err.ToStatus(
            "Requester user ID is not specified",
            "Invalid requester user ID",
        )
    }
    return nil
}

func (f *Filter) ValidateUIDs() *Error.Status {
    if err := f.ValidateTargetUID(); err != nil {
        return err
    }
    if err := f.ValidateRequesterUID(); err != nil {
        return err
    }
    return nil
}

