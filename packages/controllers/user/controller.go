package user

import (
	"log"
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/json"
	"sentinel/packages/models/role"
	"sentinel/packages/models/token"
	user "sentinel/packages/models/user"
	"sentinel/packages/net"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/mongo"
)

type Controller struct {
	user  *user.Model
	token *token.Model
}

func New(dbClient *mongo.Client) *Controller {
	return &Controller{
		user:  user.New(dbClient),
		token: token.New(dbClient),
	}
}

func (c Controller) Create(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodPost); !ok {
		return
	}

	body, ok := json.Decode[net.AuthRequestBody](req.Body, w)

	if !ok {
		net.Response.InternalServerError(w)

		return
	}

	_, err := c.user.Create(body.Email, body.Password)

	if err != nil {
		ok, e := ExternalError.Is(err)

		if !ok {
			net.Response.Message("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError, w)

			log.Fatalln(err)
		}

		net.Response.Message(e.Message, e.Status, w)

		net.Request.PrintError("Failed to create new user: "+e.Message, e.Status, req)

		return
	}

	if err := net.Response.OK(w); err != nil {
		net.Response.SendError("Failed to send success response", http.StatusInternalServerError, req, w)

		return
	}

	net.Request.Print("New user created, email: "+body.Email, req)
}

func (c Controller) UNSAFE_ChangeEmail(w http.ResponseWriter, req *http.Request) {
	net.Response.InternalServerError(w)

	log.Fatalln("[ CRITICAL ERROR ] Method not implemented")
}

func (c Controller) UNSAFE_ChangePassword(w http.ResponseWriter, req *http.Request) {
	net.Response.InternalServerError(w)

	log.Fatalln("[ CRITICAL ERROR ] Method not implemented")
}

func (c Controller) UNSAFE_ChangeRole(w http.ResponseWriter, req *http.Request) {
	net.Response.InternalServerError(w)

	log.Fatalln("[ CRITICAL ERROR ] Method not implemented")
}

func (c Controller) SoftDelete(w http.ResponseWriter, req *http.Request) {
	if ok := net.Request.Preprocessing(w, req, http.MethodDelete); !ok {
		return
	}

	accessToken, err := c.token.GetAccessToken(req)

	if err != nil {
		net.Response.SendError(err.Message, err.Status, req, w)

		return
	}

	// If token is valid, then we can trust claims
	claims := accessToken.Claims.(jwt.MapClaims)
	claimsUID := claims[token.IdKey].(string)
	claimsRole := claims[token.IdKey].(string)

	if !role.IsValid(claimsRole) {
		net.Response.SendError("Ошибка аутентификации: неверная роль, попробуйте переавторизоваться", 400, req, w)

		return
	}

	body, ok := json.Decode[net.SoftDeleteBody](req.Body, w)

	if !ok {
		if err := net.Response.InternalServerError(w); err != nil {
			panic(err)
		}
	}

	if err := c.user.SoftDelete(body.UID, claimsUID, claimsRole); err != nil {
		isExternal, e := ExternalError.Is(err)

		if !isExternal {
			net.Response.InternalServerError(w)

			return
		}

		net.Response.SendError(e.Message, e.Status, req, w)

		return
	}

	net.Response.OK(w)
}

func (c Controller) UNSAFE_HardDelete(w http.ResponseWriter, req *http.Request) {
	net.Response.InternalServerError(w)

	log.Fatalln("[ CRITICAL ERROR ] Method not implemented")
}
