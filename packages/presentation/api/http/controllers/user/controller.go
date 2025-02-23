package usercontroller

import (
	"log"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/DB"
	UserMapper "sentinel/packages/infrastructure/mappers"
	"sentinel/packages/infrastructure/token"
	datamodel "sentinel/packages/presentation/data"

	"github.com/StepanAnanin/weaver"
	"github.com/golang-jwt/jwt"
)

func newUserFilter(req *http.Request) (*UserDTO.Filter, *Error.Status) {
	body, err := getReqBody[datamodel.UidBody](req)

	if err != nil {
		return nil, err
	}

	accessToken, err := token.GetAccessToken(req)

	if err != nil {
		return nil, err
	}

	// If token is valid, then we can trust claims
	filter, err := UserMapper.FilterDTOFromClaims(body.UID, accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		return nil, err
	}

	return filter, nil
}

func getReqBody[T any](req *http.Request) (T, *Error.Status) {
	body, err := datamodel.Decode[T](req.Body)

	if err != nil {
		return body, Error.NewStatusError("Failed to decode JSON", http.StatusBadRequest)
	}

	return body, nil
}

func Create(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

	body, e := datamodel.Decode[datamodel.AuthRequestBody](req.Body)

	if e != nil {
		res.Message("Failed to decode JSON", http.StatusBadRequest)
		return
	}

	uid, err := DB.Database.Create(body.Login, body.Password)

	if err != nil {
		is, e := Error.IsStatusError(err)

		if !is {
			res.Message("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)

			log.Fatalln(err)
		}

		res.Message(e.Message, e.Status)

		return
	}

	resBody, e := datamodel.Encode(datamodel.UidBody{UID: uid})

	if e != nil {
		res.InternalServerError()
		return
	}

	res.SendBody(resBody)
}

func ChangeLogin(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

	body, err := getReqBody[datamodel.UidAndLoginBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	filter, err := newUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := DB.Database.ChangeLogin(filter, body.Login); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func ChangePassword(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

	body, err := getReqBody[datamodel.UidAndPasswordBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	filter, err := newUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := DB.Database.ChangePassword(filter, body.Password); e != nil {
		res.Message(e.Message, e.Status)
		return
	}

	res.OK()
}

func ChangeRole(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

	body, err := getReqBody[datamodel.UidAndRolesBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	filter, err := newUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if e := DB.Database.ChangeRoles(filter, body.Roles); e != nil {
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

	if e := DB.Database.SoftDelete(filter); e != nil {
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

	if e := DB.Database.Restore(filter); e != nil {
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

	if err := DB.Database.Drop(filter); err != nil {
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

	requester, err := UserMapper.PayloadFromClaims(accessToken.Claims.(jwt.MapClaims))

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	if err := DB.Database.DropAllSoftDeleted(requester.Roles); err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	res.OK()
}

func GetRole(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

	filter, err := newUserFilter(req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	roles, err := DB.Database.GetRoles(filter)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	resBody, e := datamodel.Encode(datamodel.UserRolesResponseBody{Roles: roles})

	if e != nil {
		res.InternalServerError()
		return
	}

	res.SendBody(resBody)
}

func CheckIsLoginExists(w http.ResponseWriter, req *http.Request) {
	res := weaver.NewResponse(w).Logged(req)

	body, err := getReqBody[datamodel.LoginBody](req)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	isExists, err := DB.Database.IsLoginExists(body.Login)

	if err != nil {
		res.Message(err.Message, err.Status)
		return
	}

	resBody, e := datamodel.Encode(datamodel.LoginExistanceResponseBody{Exists: isExists})

	if e != nil {
		res.InternalServerError()
		return
	}

	res.SendBody(resBody)
}
