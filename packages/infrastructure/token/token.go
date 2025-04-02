package token

import (
	"log"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/common/config"
	"sentinel/packages/common/util"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"

	Error "sentinel/packages/common/errors"
)

type SignedToken struct {
	Value string
	TTL   int64
}

type AccessToken = SignedToken
type RefreshToken = SignedToken

const RefreshTokenKey string = "refreshToken"

// UID
const IdKey string = "jti"

// Login
const IssuerKey string = "iss"

// Roles
const SubjectKey string = "sub"

func newTokenBuilder(payload *UserDTO.Payload, TTL int64) *jwt.Token {
    return jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.StandardClaims{
        IssuedAt: time.Now().Unix(),
        // For certain values see config
        ExpiresAt: TTL,
        Id:        payload.ID,
        Issuer:    payload.Login,
        Subject:   strings.Join(payload.Roles, ","),
    })
}

// Generate access and refresh tokens.
func Generate(payload *UserDTO.Payload) (*AccessToken, *RefreshToken) {
	accessTokenBuilder := newTokenBuilder(payload, generateAccessTokenTtlTimestamp())
	refreshTokenBuilder := newTokenBuilder(payload, generateRefreshTokenTtlTimestamp())

	accessTokenStr, e := accessTokenBuilder.SignedString(config.Secret.AccessTokenPrivateKey)
	refreshTokenStr, err := refreshTokenBuilder.SignedString(config.Secret.RefreshTokenPrivateKey)

	if e != nil {
		log.Fatalf("[ CRITICAL ERROR ] Failed to sign access token.\n%s", e)
	}

	if err != nil {
		log.Fatalf("[ CRITICAL ERROR ] Failed to sign refresh token.\n%s", err)
	}

	accessToken := &SignedToken{
		Value: accessTokenStr,
		TTL:   config.JWT.AccessTokenTTL().Milliseconds(),
	}

	refreshToken := &SignedToken{
		Value: refreshTokenStr,
		TTL:   config.JWT.RefreshTokenTTL().Milliseconds(),
	}

	log.Println("[ JWT ] New pair of tokens has been generated")

	return accessToken, refreshToken
}

// Retrieves and validates access token from authorization header.
//
// Returns token pointer and nil if valid and not expired token was found.
// Otherwise returns empty token pointer and error.
func GetAccessToken(authHeader string) (*jwt.Token, *Error.Status) {
	if authHeader == "" {
		return nil, unauthorized
	}

	accessTokenStr := strings.Split(authHeader, "Bearer ")[1]

    if accessTokenStr == "null" {
        return nil, invalidAccessToken
    }

	token, expired := parseAccessToken(accessTokenStr)

	if !token.Valid {
		return nil, invalidAccessToken
	}

	if expired {
		return nil, accessTokenExpired
	}

	return token, nil
}

// Retrieves and validates refresh token from auth cookie.
//
// Returns pointer to token and nil if valid and not expired token was found.
// Otherwise returns empty pointer to token and *Error.Status.
func GetRefreshToken(cookie *http.Cookie) (*jwt.Token, *Error.Status) {
	token, expired := parseRefreshToken(cookie.Value)

	if !token.Valid {
		return nil, invalidRefreshToken
	}

	if expired {
		return nil, refreshTokenExpired
	}

	return token, nil
}

func parseAccessToken(accessToken string) (*jwt.Token, bool) {
	token, _ := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return config.Secret.AccessTokenPublicKey, nil
	})

	exp := !token.Claims.(jwt.MapClaims).VerifyExpiresAt(util.UnixTimeNow(), true)

	return token, exp
}

func parseRefreshToken(refreshToken string) (*jwt.Token, bool) {
	token, _ := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return config.Secret.RefreshTokenPublicKey, nil
	})

	exp := !token.Claims.(jwt.MapClaims).VerifyExpiresAt(util.UnixTimeNow(), true)

	return token, exp
}

