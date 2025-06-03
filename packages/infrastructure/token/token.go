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

var tokenLogger = logger.NewSource("TOKEN", logger.Default)

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

const (
    ServiceIdClaimsKey string = "iss"
    UserIdClaimsKey string = "sub"
    IssuedAtClaimsKey string = "iat"
    ExpiresAtClaimsKey string = "exp"
    UserRolesClaimsKey string = "roles"
    UserLoginClaimsKey string = "login"
)

type Claims struct {
    Roles []string `json:"roles"`
    Login string `json:"login"`

    jwt.RegisteredClaims
}

func newSignedToken(
    payload *UserDTO.Payload,
    ttl time.Duration,
    key ed25519.PrivateKey,
) (*SignedToken, *Error.Status) {
    now := jwt.NewNumericDate(time.Now())
    claims := Claims{
        Login: payload.Login,
        Roles: payload.Roles,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    config.App.ServiceID,
            IssuedAt:  now,
            NotBefore: now,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl).UTC()),
            Subject:   payload.ID,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

    tokenStr, err := token.SignedString(key)
    if err != nil {
        tokenLogger.Error("Failed to sign token", err.Error(), nil)
        return nil, Error.StatusInternalError
    }

    return &SignedToken{tokenStr, ttl.Milliseconds()}, nil
}

func NewAccessToken(payload *UserDTO.Payload) (*SignedToken, *Error.Status) {
    return newSignedToken(
        payload,
        config.Auth.AccessTokenTTL(),
        config.Secret.AccessTokenPrivateKey,
    )
}

func NewRefreshToken(payload *UserDTO.Payload) (*SignedToken, *Error.Status) {
    return newSignedToken(
        payload,
        config.Auth.RefreshTokenTTL(),
        config.Secret.RefreshTokenPrivateKey,
    )
}

type AccessToken = SignedToken
type RefreshToken = SignedToken

// Both tokens types are aliases for token.SignedToken
func NewAuthTokens(payload *UserDTO.Payload) (*AccessToken, *RefreshToken, *Error.Status) {
    var (
        atk *SignedToken
        rtk *SignedToken
        err *Error.Status
    )

    atk, err = NewAccessToken(payload)
    if err != nil {
        return nil, nil, err
    }

    rtk, err = NewRefreshToken(payload)
    if err != nil {
        return nil, nil, err
    }

    return atk, rtk, nil
}

func NewActivationToken(uid string, login string, roles []string) (*SignedToken, *Error.Status) {
    return newSignedToken(
        &UserDTO.Payload{
            ID: uid,
            Roles: roles,
        },
        config.App.ActivationTokenTTL(),
        config.Secret.ActivationTokenPrivateKey,
    )
}

// Parses and validates given token
func ParseSingedToken(tokenStr string, key ed25519.PublicKey) (*jwt.Token, *Error.Status) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
        if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
		return key, nil
	})
    if err != nil {
        switch {
        case errors.Is(err, jwt.ErrTokenMalformed):
            return nil, TokenMalformed
        case errors.Is(err, jwt.ErrTokenExpired):
            return nil, TokenExpired
        case errors.Is(err, jwt.ErrTokenNotValidYet):
            // Will never trigger for our current tokens since we don't set NBF
            return nil, TokenModified
        case errors.Is(err, jwt.ErrTokenSignatureInvalid):
            // Check if someone tampered with the token
            return nil, TokenModified
        default:
            tokenLogger.Error("Failed to parse signed token", err.Error(), nil)
            return nil, Error.StatusInternalError
        }
    }

	return token, nil
}

