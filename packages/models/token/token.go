package token

import (
	"errors"
	"log"
	"net/http"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/role"
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
func (m *Model) Generate(user *user.Payload) (*SignedToken, *SignedToken) {
	accessTokenBuilder := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.StandardClaims{
		IssuedAt: time.Now().Unix(),
		// For certain values see config
		ExpiresAt: generateAccessTokenTtlTimestamp(),
		Id:        user.ID,
		Issuer:    user.Email,
		Subject:   string(user.Role),
	})

	refreshTokenBuilder := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.StandardClaims{
		IssuedAt: time.Now().Unix(),
		// For certain values see config
		ExpiresAt: generateRefreshTokenTtlTimestamp(),
		Id:        user.ID,
		Issuer:    user.Email,
		Subject:   string(user.Role),
	})

	accessTokenStr, e := accessTokenBuilder.SignedString(*config.JWT.AccessTokenPrivateKey)
	refreshTokenStr, err := refreshTokenBuilder.SignedString(*config.JWT.RefreshTokenPrivateKey)

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
func (m *Model) GetAccessToken(req *http.Request) (*jwt.Token, *ExternalError.Error) {
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

	return token, nil
}

// Retrieves and validates refresh token from request.
//
// Returns token pointer and nil if valid and not expired token was found.
// Otherwise returns empty token pointer and error, this error is either http.ErrNoCookie, either ExternalError.Error
func (m *Model) GetRefreshToken(req *http.Request) (*jwt.Token, error) {
	var emptyToken *jwt.Token

	authCookie, err := req.Cookie(RefreshTokenKey)

	if err != nil {
		// If this condition is true, that mean error ocured inside of "req.Cookie(...)"
		if !errors.Is(err, http.ErrNoCookie) {
			log.Fatalln(err)
		}

		return emptyToken, err
	}

	token, expired := m.parseRefreshToken(authCookie.Value)

	if !token.Valid {
		return emptyToken, ExternalError.New("Invalid refresh token", http.StatusBadRequest)
	}

	if expired {
		// TODO
		// Not sure that status 409 is OK for this case, currently this tells user that there are conflict with server and him,
		// and reason of conflict in next: User assumes that he authorized but it's wrong, cuz refresh token expired.
		// More likely will be better to use status 401 (unathorized) in this case, but once againg - i'm not sure.
		return emptyToken, ExternalError.New("Refresh token expired", http.StatusConflict)
	}

	return token, nil
}

func (m *Model) parseAccessToken(accessToken string) (*jwt.Token, bool) {
	token, _ := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return *config.JWT.AccessTokenPublicKey, nil
	})

	exp := !token.Claims.(jwt.MapClaims).VerifyExpiresAt(util.UnixTimeNow(), true)

	return token, exp
}

func (m *Model) parseRefreshToken(refreshToken string) (*jwt.Token, bool) {
	token, _ := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return *config.JWT.RefreshTokenPublicKey, nil
	})

	exp := !token.Claims.(jwt.MapClaims).VerifyExpiresAt(util.UnixTimeNow(), true)

	return token, exp
}

// IMPORTANT: Use this function only if token is valid.
// TODO Return error istead of crushing app
func (m *Model) PayloadFromClaims(claims jwt.MapClaims) (*user.Payload, *ExternalError.Error) {
	var r *user.Payload

	if err := verifyClaims(claims); err != nil {
		return r, err
	}

	return &user.Payload{
		ID:    claims[IdKey].(string),
		Email: claims[IssuerKey].(string),
		Role:  role.Role(claims[SubjectKey].(string)),
	}, nil
}

func (m *Model) UserFilterFromClaims(targetUID string, claims jwt.MapClaims) (*user.Filter, *ExternalError.Error) {
	var r *user.Filter

	if err := verifyClaims(claims); err != nil {
		return r, err
	}

	return &user.Filter{
		TargetUID:     targetUID,
		RequesterUID:  claims[IdKey].(string),
		RequesterRole: role.Role(claims[SubjectKey].(string)),
	}, nil
}
