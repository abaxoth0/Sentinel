package usercontroller

import (
	"log"
	"net/http"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/json"
	"sentinel/packages/models/token"
	user "sentinel/packages/models/user"

	"github.com/StepanAnanin/weaver/http/response"
	"github.com/golang-jwt/jwt"
)

func buildUserFilterAndReqBody[T any](req *http.Request) (*user.Filter, T, *ExternalError.Error) {
	var emptyFilter *user.Filter
	var emptyReqBody T

	rawBody, ok := json.Decode[any](req.Body)

	if !ok {
		return emptyFilter, emptyReqBody, ExternalError.New("Failed to decode JSON", http.StatusBadRequest)
	}

	accessToken, err := token.GetAccessToken(req)

	if err != nil {
		return emptyFilter, emptyReqBody, err
	}

	body, _ := rawBody.(json.UidBody)

	// If token is valid, then we can trust claims
	filter, err := token.UserFilterFromClaims(body.UID, accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		return emptyFilter, emptyReqBody, err
	}

	if err := filter.RequesterRole.Verify(); err != nil {
		return emptyFilter, emptyReqBody, err
	}

	return filter, rawBody.(T), nil
}

func Create(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)
	body, ok := json.Decode[json.AuthRequestBody](req.Body)

	if !ok {
		res.Message("Failed to decode JSON", http.StatusBadRequest)
		return
	}

	_, err := user.Create(body.Login, body.Password)

	if err != nil {
		ok, e := ExternalError.Is(err)

		if !ok {
			res.Message("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)

			log.Fatalln(err)
		}

		res.Message(e.Message, e.Status)

		return
	}

	res.OK()
}

func ChangeLogin(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)
	filter, body, err := buildUserFilterAndReqBody[json.UidAndLoginBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
	}

	if e := user.ChangeLogin(filter, body.Login); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func ChangePassword(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)
	filter, body, err := buildUserFilterAndReqBody[json.UidAndPasswordBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := user.ChangePassword(filter, body.Password); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func ChangeRole(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)
	filter, body, err := buildUserFilterAndReqBody[json.UidAndRoleBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := user.ChangeRole(filter, body.Role); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func SoftDelete(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)
	filter, _, err := buildUserFilterAndReqBody[json.UidBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := user.SoftDelete(filter); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func Restore(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)
	filter, _, err := buildUserFilterAndReqBody[json.UidBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := user.Restore(filter); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

// Hard delete
func Drop(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)
	filter, _, err := buildUserFilterAndReqBody[json.UidBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if err := user.Drop(filter); err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	res.OK()
}

func DropAllDeleted(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)

	accessToken, err := token.GetAccessToken(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	requester, err := token.PayloadFromClaims(accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if err := user.DropAllDeleted(requester.Role); err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	res.OK()
}

func GetRole(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)
	filter, _, err := buildUserFilterAndReqBody[json.UidBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	role, err := user.GetRole(filter)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	resBody, ok := json.Encode(json.UserRoleResponseBody{Role: role})

	if !ok {
		res.InternalServerError()
		return
	}

	res.SendBody(resBody)
}

func CheckIsLoginExists(w http.ResponseWriter, req *http.Request) {
	res := response.New(w).Logged(req)
	body, ok := json.Decode[json.LoginBody](req.Body)

	if !ok {
		res.Message("Failed to decode JSON", http.StatusBadRequest)
		return
	}

	isExists, err := user.CheckIsLoginExists(body.Login)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	resBody, ok := json.Encode(json.LoginExistanceResponseBody{Exists: isExists})

	if !ok {
		res.InternalServerError()
		return
	}

	res.SendBody(resBody)
}
