package token

import (
	"crypto/ed25519"
	"fmt"
	"log"
	"sentinel/packages/common/config"
	"sentinel/packages/common/util"
	UserDTO "sentinel/packages/core/user/DTO"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"

	Error "sentinel/packages/common/errors"
)

// TODO currently claims used in a wrong way, need to fix that
//      (e.g. ISS must contain this service id instead of user login)
const (
    // UID
    IdKey string = "jti"
    // Login
    IssuerKey string = "iss"
    // Roles
    SubjectKey string = "sub"
)

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

func newSignedToken(
    payload *UserDTO.Payload,
    ttl time.Duration,
    key ed25519.PrivateKey,
) (*SignedToken, *Error.Status) {
    token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.StandardClaims{
        IssuedAt: time.Now().Unix(),
        // For certain values see config
        ExpiresAt: util.TimestampSinceNow(ttl),
        Id:        payload.ID,
        Issuer:    payload.Login,
        Subject:   strings.Join(payload.Roles, ","),
    })

    tokenStr, err := token.SignedString(key)
    if err != nil {
        log.Printf("[ JWT ] Failed to sign token: %s\n", err.Error())
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

// Parses and validates given token
func ParseSingedToken(tokenStr string, key ed25519.PublicKey) (*jwt.Token, *Error.Status) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
        if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
		return key, nil
	})
    if err != nil {
        ve, ok := err.(*jwt.ValidationError)
        if !ok {
            log.Printf("[ UNKNOWN ERROR ] Failed to parse signed token: %s\n", err.Error())
            return nil, Error.StatusInternalError
        }

        switch {
        case ve.Errors & jwt.ValidationErrorMalformed != 0:
            return nil, TokenMalformed
        case ve.Errors & jwt.ValidationErrorExpired != 0:
            return nil, TokenExpired
        case ve.Errors & jwt.ValidationErrorNotValidYet != 0:
            // Will never trigger for our current tokens since we don't set NBF
            return nil, TokenModified
        case ve.Errors & jwt.ValidationErrorSignatureInvalid != 0:
            // Check if someone tampered with the token
            return nil, TokenModified
        default:
            log.Printf("[ UNKNOWN ERROR ] Failed to parse signed token: %s\n", err.Error())
            return nil, InvalidToken
        }
    }

    exp := !token.Claims.(jwt.MapClaims).VerifyExpiresAt(util.UnixTimeNow(), true)
    if exp {
        return nil, TokenExpired
    }

	return token, nil
}

