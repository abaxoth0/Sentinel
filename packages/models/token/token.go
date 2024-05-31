package token

import (
	"log"
	"net/http"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/user"
	"sentinel/packages/util"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/mongo"
)

// There are no need to store tokens in db, doing so will
// just cause problems with auth on multiple devices.
// (https://stackoverflow.com/questions/73257330/multiple-device-login-using-jwt)
type Model struct{}

func New(dbClient *mongo.Client) *Model {
	return &Model{}
}

type SignedToken struct {
	Value string
	TTL   int64
}

const RefreshTokenKey string = "refreshToken"

// UID
const IdKey string = "jti"

// E-Mail
const IssuerKey string = "iss"

// Role
const SubjectKey string = "sub"

// Generate access and refresh tokens. (they returns in same order as here)
func (m Model) Generate(user user.Payload) (SignedToken, SignedToken) {
	accessTokenBuilder := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.StandardClaims{
		IssuedAt: time.Now().Unix(),
		// For certain values see config
		ExpiresAt: generateAccessTokenTtlTimestamp(),
		Id:        user.ID,
		Issuer:    user.Email,
		Subject:   user.Role,
	})

	refreshTokenBuilder := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.StandardClaims{
		IssuedAt: time.Now().Unix(),
		// For certain values see config
		ExpiresAt: generateRefreshTokenTtlTimestamp(),
		Id:        user.ID,
		Issuer:    user.Email,
		Subject:   user.Role,
	})

	accessTokenStr, e := accessTokenBuilder.SignedString(*config.JWT.AccessTokenPrivateKey)
	refreshTokenStr, err := refreshTokenBuilder.SignedString(*config.JWT.RefreshTokenPrivateKey)

	if e != nil {
		log.Fatalf("[ CRITICAL ERROR ] Failed to sign access token.\n%s", e)
	}

	if err != nil {
		log.Fatalf("[ CRITICAL ERROR ] Failed to sign refresh token.\n%s", err)
	}

	accessToken := SignedToken{
		Value: accessTokenStr,
		TTL:   config.JWT.AccessTokenTTL.Milliseconds(),
	}

	refreshToken := SignedToken{
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
func (m Model) GetAccessToken(req *http.Request) (*jwt.Token, *ExternalError.Error) {
	var r *jwt.Token

	authHeaderValue := req.Header.Get("Authorization")

	if authHeaderValue == "" {
		return r, ExternalError.New("Вы не авторизованы", 401)
	}

	accessTokenStr := strings.Split(authHeaderValue, "Bearer ")[1]

	token, expired := m.parseAccessToken(accessTokenStr)

	if !token.Valid {
		return r, ExternalError.New("Invalid access token", http.StatusBadRequest)
	}

	if expired {
		return r, ExternalError.New("Access token expired", http.StatusUnauthorized)
	}

	return r, nil
}

func (m Model) parseAccessToken(accessToken string) (*jwt.Token, bool) {
	token, _ := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return *config.JWT.AccessTokenPublicKey, nil
	})

	exp := !token.Claims.(jwt.MapClaims).VerifyExpiresAt(util.UnixTimeNow(), true)

	return token, exp
}

func (m Model) ParseRefreshToken(refreshToken string) (*jwt.Token, bool) {
	token, _ := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return *config.JWT.RefreshTokenPublicKey, nil
	})

	exp := !token.Claims.(jwt.MapClaims).VerifyExpiresAt(util.UnixTimeNow(), true)

	return token, exp
}

// IMPORTANT: Use this function only if token is valid.
func (m Model) PayloadFromClaims(claims jwt.MapClaims) user.Payload {
	if claims[IdKey] == nil {
		log.Fatalln("[ CRITICAL ERROR ] Malfunction token claims: \"jti\" is nil. Ensure that token is valid.")
	}

	if claims[IssuerKey] == nil {
		log.Fatalln("[ CRITICAL ERROR ] Malfunction token claims: \"iss\" is nil. Ensure that token is valid.")
	}

	if claims[SubjectKey] == nil {
		log.Fatalln("[ CRITICAL ERROR ] Malfunction token claims: \"sub\" is nil. Ensure that token is valid.")
	}

	return user.Payload{
		ID:    claims[IdKey].(string),
		Email: claims[IssuerKey].(string),
		Role:  claims[SubjectKey].(string),
	}
}
