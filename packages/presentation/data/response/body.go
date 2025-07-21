package responsebody

import (
	"sentinel/packages/core/location/DTO"
	"sentinel/packages/core/session/DTO"
)

// swagger:model TokenResponse
type Token struct {
	Message     string `json:"message" example:"hello"`
	AccessToken string `json:"accessToken" example:"eyJhbGciOi..."`
	ExpiresIn   int    `json:"expiresIn" example:"600"`
}

// swagger:model MessageResponse
type Message struct {
	Message string `json:"message" example:"message text"`
}

// swagger:model IsLoginAvailableResponse
type IsLoginAvailable struct {
	Available bool `json:"available" example:"true"`
}

// swagger:model ErrorResponse
type Error struct {
	Error 	string `json:"error" example:"Error"`
	Message string `json:"message" example:"Something went wrong"`
}

type (
	Session  = *sessiondto.Public
	Location = *locationdto.Public
)

type UserSession struct {
	Session  `json:",inline"`
	Location `json:"location"`
}

