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

type Introspection struct {
	Active 		bool		`json:"active" example:"true"`
	SessionID	string		`json:"jti" example:"ade1cdb0-309c-48c5-8251-c3f39ec0d606"`
	Subject 	string 		`json:"sub" example:"c9fcc8e3-f4f1-4b85-a65e-29bb889cbccb"`
	Issuer		string		`json:"iss" example:"3c23ebbd-42af-47c6-9c50-7295b3ac3a62"`
	Audience	[]string 	`json:"aud" example:"urn:api:auth,urn:api:billing"`
	ExpiresAt	int64		`json:"exp" example:"1753707388"`
	IssuedAt	int64		`json:"iat" example:"1753706788"`
	Scope		[]string 	`json:"scope" example:"read,write"`
}

type JSONWebKey struct {
	Kty string `json:"kty" example:"OKP"`
	Alg string `json:"alg" example:"EdDSA"`
	Kid string `json:"kid" example:"access-1"`
	Use string `json:"use" example:"sig"`
	Crv string `json:"crv,omitempty" example:"Ed25519"`
	X 	string `json:"x,omitempty" example:"Vzu3AwphVg7zmlrmAojBswMl4xoEIzsc9BY5DHGgUzo"`
}

type JWKs struct {
	Keys []JSONWebKey `json:"keys"`
}

type CSRF struct {
	Token string `json:"csrf-token" example:"h3PCI++3T0fEphsWoupOQyIQjlOx953bF0wlhMNu1jw="`
}

