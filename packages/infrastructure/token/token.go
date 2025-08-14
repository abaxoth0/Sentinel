package token

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"sentinel/packages/common/config"
	"sentinel/packages/common/logger"
	UserDTO "sentinel/packages/core/user/DTO"
	"time"

	"github.com/golang-jwt/jwt/v5"

	Error "sentinel/packages/common/errors"
)

var log = logger.NewSource("TOKEN", logger.Default)

// Must match RFC 9068
// (https://datatracker.ietf.org/doc/rfc9068/)

type SignedToken struct {
    value string
    ttl   int64
}

func (t *SignedToken) String() string {
    return t.value
}

func (t *SignedToken) TTL() int64 {
    return t.ttl
}

type tokenHeaders = map[string]string

const (
    SessionIdClaimsKey 	= "jti"
    ServiceIdClaimsKey 	= "iss"
    UserIdClaimsKey 	= "sub"
    IssuedAtClaimsKey 	= "iat"
    ExpiresAtClaimsKey 	= "exp"
	AudienceClaimsKey	= "aud"
    UserRolesClaimsKey 	= "roles"
    UserLoginClaimsKey 	= "login"
	VersionClaimsKey 	= "version"
)

type Claims struct {
    Roles 		[]string `json:"roles"`
    Login 		string 	 `json:"login"`
	Version 	uint32 	 `json:"version"`

    jwt.RegisteredClaims
}

var audienceLookup map[string]struct{}
var isInit = false

func Init() {
	if isInit {
		log.Panic("Failed to initialize Token module", "Token module already initialized", nil)
	}

	log.Info("Initializing...", nil)

    audienceLookup = make(map[string]struct{})
    for _, aud := range config.Auth.TokenAudience {
        audienceLookup[aud] = struct{}{}
    }

	log.Info("Initializing: OK", nil)

	isInit = true
}

func newSignedToken(
    payload *UserDTO.Payload,
    ttl time.Duration,
    key ed25519.PrivateKey,
	audience []string,
	headers tokenHeaders,
) (*SignedToken, *Error.Status) {
	if audience == nil || len(audience) == 0 {
		log.Error("Failed to create signed token", TokenAudienceIsNotSpecified.Error(), nil)
		return nil, TokenAudienceIsNotSpecified
	}
	for _, aud := range audience {
		if _, exists := audienceLookup[aud]; !exists {
			log.Error(
				"Failed to create signed token",
				"Audience doesn't exists: " + aud,
				nil,
			)
			return nil, TokenAudienceDoesNotExists
		}
	}

    now := jwt.NewNumericDate(time.Now())
    claims := Claims{
        Login: payload.Login,
        Roles: payload.Roles,
		Version: payload.Version,
        RegisteredClaims: jwt.RegisteredClaims{
			ID: payload.SessionID,
            Issuer:    config.App.ServiceID,
            IssuedAt:  now,
            NotBefore: now,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl).UTC()),
            Subject:   payload.ID,
			Audience: jwt.ClaimStrings(audience),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	for key := range headers {
		token.Header[key] = headers[key]
	}

    tokenStr, err := token.SignedString(key)
    if err != nil {
        log.Error("Failed to sign token", err.Error(), nil)
        return nil, Error.StatusInternalError
    }

    return &SignedToken{tokenStr, ttl.Milliseconds()}, nil
}

func NewAccessToken(payload *UserDTO.Payload) (*SignedToken, *Error.Status) {
	log.Trace("Creating new access token...", nil)

	token, err := newSignedToken(
        payload,
        config.Auth.AccessTokenTTL(),
        config.Secret.AccessTokenPrivateKey,
		payload.Audience,
		tokenHeaders{
			"typ": "at+jwt",
		},
    )
	if err != nil {
		return nil, err
	}

	log.Trace("Creating new access token: OK", nil)

	return token, nil
}

func NewRefreshToken(payload *UserDTO.Payload) (*SignedToken, *Error.Status) {
	log.Trace("Creating new refresh token...", nil)

	token, err := newSignedToken(
        payload,
        config.Auth.RefreshTokenTTL(),
        config.Secret.RefreshTokenPrivateKey,
		[]string{config.Auth.SelfAudience},
		nil,
    )
	if err != nil {
		return nil, err
	}

	log.Trace("Creating new refresh token: OK", nil)

	return token, nil
}

func NewAuthTokens(payload *UserDTO.Payload) (accessToken *SignedToken, refreshToken *SignedToken, err *Error.Status) {
    accessToken, err = NewAccessToken(payload)
    if err != nil {
        return nil, nil, err
    }

    refreshToken, err = NewRefreshToken(payload)
    if err != nil {
        return nil, nil, err
    }

    return accessToken, refreshToken, nil
}

func NewActivationToken(uid string, login string) (*SignedToken, *Error.Status) {
	log.Trace("Creating new activation token...", nil)

	token, err := newSignedToken(
        &UserDTO.Payload{
            ID: uid,
			Login: login,
        },
        config.App.ActivationTokenTTL(),
        config.Secret.ActivationTokenPrivateKey,
		[]string{config.Auth.SelfAudience},
		tokenHeaders{
			"typ": "activation",
		},
    )
	if err != nil {
		return nil, err
	}

	log.Trace("Creating new activation token: OK", nil)

	return token, nil
}

func NewPasswordResetToken(uid string, login string) (*SignedToken, *Error.Status) {
	log.Trace("Creating new password reset token...", nil)

	token, err := newSignedToken(
		&UserDTO.Payload{
			ID: uid,
			Login: login,
		},
        config.App.PasswordResetTokenTTL(),
        config.Secret.PasswordResetTokenPrivateKey,
		[]string{config.Auth.SelfAudience},
		tokenHeaders{
			"typ": "password-reset",
		},
    )
	if err != nil {
		return nil, err
	}

	log.Trace("Creating new password reset token: OK", nil)

	return token, nil
}

var jwtParserOptions = []jwt.ParserOption{
	jwt.WithLeeway(5 * time.Second),
}

func ed25519KeyFunc(key ed25519.PublicKey) func (token *jwt.Token) (any, error) {
	return func(token *jwt.Token) (any, error) {
		// RFC 9068 p2.1
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	}
}

// Parses and validates given token
func ParseSingedToken(tokenStr string, key ed25519.PublicKey) (*jwt.Token, *Error.Status) {
	log.Trace("Parsing signed token...", nil)

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, ed25519KeyFunc(key), jwtParserOptions...)
    if err != nil {
		var e *Error.Status

        switch {
        case errors.Is(err, jwt.ErrTokenMalformed):
            e = TokenMalformed
        case errors.Is(err, jwt.ErrTokenExpired):
            e = TokenExpired
        case errors.Is(err, jwt.ErrTokenNotValidYet):
            e = TokenModified
        case errors.Is(err, jwt.ErrTokenSignatureInvalid):
            e = TokenInvalidSignature
        default:
            log.Error("Failed to parse signed token", err.Error(), nil)
            return nil, Error.StatusInternalError
        }

		log.Error("Invalid token", e.Error(), nil)

		return nil, e
    }

	if token.Header["typ"] == "activation" || token.Header["typ"] == "password-reset" {
		log.Trace("Parsing signed token: OK", nil)
		return token, nil
	}

	// RFC 9068 p2.2
	if claims.Issuer == "" ||
		claims.Subject == "" ||
		len(claims.Audience) == 0 ||
		claims.ExpiresAt == nil ||
		claims.IssuedAt == nil ||
		claims.ID == "" {
		log.Error("Token validation failed", TokenMissingRequiredClaims.Error(), nil)
		return nil, TokenMissingRequiredClaims
	}

	for _, tokenAud := range claims.Audience {
		if _, exists := audienceLookup[tokenAud]; !exists {
			log.Error("Token validation failed", TokenAudienceMismatch.Error(), nil)
			return nil, TokenAudienceMismatch
		}
	}

	log.Trace("Parsing signed token: OK", nil)

	return token, nil
}

