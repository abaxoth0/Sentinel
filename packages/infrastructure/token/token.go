package token

import (
	"errors"
	"log"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/config"
	"sentinel/packages/util"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"

	Error "sentinel/packages/errors"
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

// Generate access and refresh tokens.
func Generate(payload *UserDTO.Payload) (*AccessToken, *RefreshToken) {
	accessTokenBuilder := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.StandardClaims{
		IssuedAt: time.Now().Unix(),
		// For certain values see config
		ExpiresAt: generateAccessTokenTtlTimestamp(),
		Id:        payload.ID,
		Issuer:    payload.Login,
		Subject:   strings.Join(payload.Roles, ","),
	})

	refreshTokenBuilder := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.StandardClaims{
		IssuedAt: time.Now().Unix(),
		// For certain values see config
		ExpiresAt: generateRefreshTokenTtlTimestamp(),
		Id:        payload.ID,
		Issuer:    payload.Login,
		Subject:   strings.Join(payload.Roles, ","),
	})

	accessTokenStr, e := accessTokenBuilder.SignedString(config.JWT.AccessTokenPrivateKey)
	refreshTokenStr, err := refreshTokenBuilder.SignedString(config.JWT.RefreshTokenPrivateKey)

	if e != nil {
		log.Fatalf("[ CRITICAL ERROR ] Failed to sign access token.\n%s", e)
	}

	if err != nil {
		log.Fatalf("[ CRITICAL ERROR ] Failed to sign refresh token.\n%s", err)
	}

	accessToken := &SignedToken{
		Value: accessTokenStr,
		TTL:   config.JWT.AccessTokenTTL.Milliseconds(),
	}

	refreshToken := &SignedToken{
		Value: refreshTokenStr,
		TTL:   config.JWT.RefreshTokenTTL.Milliseconds(),
	}

	log.Println("[ JWT ] New pair of tokens has been generated")

	return accessToken, refreshToken
}

// Retrieves and validates access token from request.
//
// Returns token pointer and nil if valid and not expired token was found.
// Otherwise returns empty token pointer and error.
func GetAccessToken(req *http.Request) (*jwt.Token, *Error.Status) {
    var r *jwt.Token

	authHeaderValue := req.Header.Get("Authorization")

	if authHeaderValue == "" {
		return r, unauthorized
	}

	accessTokenStr := strings.Split(authHeaderValue, "Bearer ")[1]

    if accessTokenStr == "null" {
        return r, invalidAccessToken
    }

	token, expired := parseAccessToken(accessTokenStr)

	if !token.Valid {
		return r, invalidAccessToken
	}

	if expired {
		return r, accessTokenExpired
	}

	return token, nil
}

// Retrieves and validates refresh token from request.
//
// Returns token pointer and nil if valid and not expired token was found.
// Otherwise returns empty token pointer and error, this error is either http.ErrNoCookie, either ExternalError.Error
func GetRefreshToken(req *http.Request) (*jwt.Token, error) {
	var emptyToken *jwt.Token

	authCookie, err := req.Cookie(RefreshTokenKey)

	if err != nil {
		// If this condition is true, that mean error ocured inside of "req.Cookie(...)"
		if !errors.Is(err, http.ErrNoCookie) {
			log.Fatalln(err)
		}

		return emptyToken, err
	}

	token, expired := parseRefreshToken(authCookie.Value)

	if !token.Valid {
		return emptyToken, invalidRefreshToken
	}

	if expired {
		return emptyToken, refreshTokenExpired
	}

	return token, nil
}

func parseAccessToken(accessToken string) (*jwt.Token, bool) {
	token, _ := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return config.JWT.AccessTokenPublicKey, nil
	})

	exp := !token.Claims.(jwt.MapClaims).VerifyExpiresAt(util.UnixTimeNow(), true)

	return token, exp
}

func parseRefreshToken(refreshToken string) (*jwt.Token, bool) {
	token, _ := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return config.JWT.RefreshTokenPublicKey, nil
	})

	exp := !token.Claims.(jwt.MapClaims).VerifyExpiresAt(util.UnixTimeNow(), true)

	return token, exp
}
