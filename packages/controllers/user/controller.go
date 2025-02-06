package usercontroller

import (
	"log"
	"net/http"
	"sentinel/packages/entities"
	Error "sentinel/packages/errs"
	"sentinel/packages/json"
	"sentinel/packages/models/token"
	user "sentinel/packages/models/user"

	"github.com/StepanAnanin/weaver"
	"github.com/golang-jwt/jwt"
)

func newUserFilter(req *http.Request) (*entities.UserFilter, *Error.HTTP) {
    body, err := getReqBody[json.UidBody](req);

    if err != nil {
		return nil, err
	}

    accessToken, err := token.GetAccessToken(req);

    if err != nil {
		return nil, err
	}

	// If token is valid, then we can trust claims
	filter, err := user.NewFilterFromClaims(body.UID, accessToken.Claims.(jwt.MapClaims))

    if err != nil {
		return nil, err
	}

	return filter, nil
}

func getReqBody[T interface{}](req *http.Request) (T, *Error.HTTP) {
    body, ok := json.Decode[T](req.Body);

    if !ok {
		return body, Error.NewHTTP("Failed to decode JSON", http.StatusBadRequest)
    }

    return body, nil
}

func Create(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)
    body, ok := json.Decode[json.AuthRequestBody](req.Body)

    if !ok {
		res.Message("Failed to decode JSON", http.StatusBadRequest)
		return
	}

	uid, err := user.Create(body.Login, body.Password)

    if err != nil {
		ok, e := Error.Is(err)

		if !ok {
			res.Message("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)

			log.Fatalln(err)
		}

		res.Message(e.Message, e.Status)

		return
	}

	resBody, ok := json.Encode(json.UidBody{UID: uid.Hex()})

    if !ok {
		res.InternalServerError()
		return
	}

	res.SendBody(resBody)
}

func ChangeLogin(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

    body, err := getReqBody[json.UidAndLoginBody](req);

    if err != nil {
        res.Message(err.Message, err.Status)
        return
    }

    filter, err := newUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)
	    return
    }

	if e := user.ChangeLogin(filter, body.Login); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func ChangePassword(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

    body, err := getReqBody[json.UidAndPasswordBody](req);

    if err != nil {
        res.Message(err.Message, err.Status);
        return;
    }

    filter, err := newUserFilter(req)

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
	res := weaver.NewResponse(w).Logged(req)

    body, err := getReqBody[json.UidAndRoleBody](req);

    if err != nil {
        res.Message(err.Message, err.Status);
        return;
    }

    filter, err := newUserFilter(req)

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
	res := weaver.NewResponse(w).Logged(req)

    filter, err := newUserFilter(req)

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
	res := weaver.NewResponse(w).Logged(req)

    filter, err := newUserFilter(req)

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
	res := weaver.NewResponse(w).Logged(req)

    filter, err := newUserFilter(req)

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
	res := weaver.NewResponse(w).Logged(req)

	accessToken, err := token.GetAccessToken(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	requester, err := user.PayloadFromClaims(accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if err := user.DropAllDeleted(requester.Roles); err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	res.OK()
}

// TODO fix, now cause panic
func GetRole(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

    filter, err := newUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	roles, err := user.GetRoles(filter)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	resBody, ok := json.Encode(json.UserRoleResponseBody{Roles: roles})

	if !ok {
		res.InternalServerError()
		return
	}

	res.SendBody(resBody)
}

func CheckIsLoginExists(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

    body, err := getReqBody[json.LoginBody](req);

    if err != nil {
        res.Message(err.Message, err.Status);
        return;
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
