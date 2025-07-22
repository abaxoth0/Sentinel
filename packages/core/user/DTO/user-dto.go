package userdto

import (
	"slices"
	"time"
)

/*
	!!!!! ACHTUNG !!!!!
	If you will change any of this DTOs then don't forget to change protobuf models, (./packages/common/proto/user.proto)
	cuz they are used for cache, so if you won't updated them this will lead to cache data loss and may lead to undefined behaviour.
	Also don't forget to update protobuf encoder (./packages/common/encoder/protobuf.go).
*/

type Any interface {
    IsDeleted() bool
}

type Public struct {
    ID           string     `json:"id"`
	Login        string     `json:"login"`
	Roles        []string   `json:"roles"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty"`
	Version 	 uint32	    `json:"version"`
}

func (dto *Public) IsDeleted() bool {
    return !dto.DeletedAt.IsZero()
}

func (dto *Public) IsActive() bool {
    return !slices.Contains(dto.Roles, "unconfirmed_user")
}

type Basic struct {
    ID           string    `json:"id"`
	Login        string    `json:"login"`
	Password     string    `json:"password"`
	Roles        []string  `json:"roles"`
	DeletedAt    time.Time `json:"deletedAt"`
	Version 	 uint32	   `json:"version"`
}

// Creates new copy of this DTO, returns non-nil pointer to it
func (dto *Basic) Copy() *Basic {
	roles := make([]string, len(dto.Roles))
	copy(roles, dto.Roles)
	return &Basic{
		ID: dto.ID,
		Login: dto.Login,
		Password: dto.Password,
		Roles: roles,
		DeletedAt: dto.DeletedAt,
	}
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
	Version 	 uint32	   `json:"version"`
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
    ID               string    	`json:"id"`
    ChangedUserID    string    	`json:"changedUserID"`
    ChangedByUserID  string    	`json:"changedByUserID"`
    Operation        string    	`json:"operation"`
    Login            string    	`json:"login"`
	Password         string    	`json:"password"`
	Roles            []string  	`json:"roles"`
	DeletedAt        time.Time 	`json:"deletedAt"`
    ChangedAt        time.Time 	`json:"changedAt"`
	Version 	 	 uint32	   	`json:"version"`
	Reason			 string		`json:"reason,omitempty"`
}

func (dto *Audit) IsDeleted() bool {
    return !dto.DeletedAt.IsZero()
}

// swagger:model UserPayload
type Payload struct {
	ID    		string   `json:"id" example:"d529a8d2-1eb4-4bce-82aa-e62095dbc653"`
	Login 		string   `json:"login" example:"admin@mail.com"`
	Roles 		[]string `json:"roles" example:"user,moderator"`
	Version 	uint32	 `json:"version" example:"7"`
	SessionID 	string 	 `json:"session-id" example:"35b92582-7694-4958-9751-1fef710cb94d"`
}

