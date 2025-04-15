package usercontroller

import (
	"errors"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	"sentinel/packages/core/user"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authentication"
	"sentinel/packages/infrastructure/email"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	controller "sentinel/packages/presentation/api/http/controllers"
	datamodel "sentinel/packages/presentation/data"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func getUserIdFromPath(ctx echo.Context) (string, *echo.HTTPError) {
    uid := ctx.Param("uid")

    if err := validation.UUID(uid); err != nil {
        if err == Error.NoValue {
            return "", echo.NewHTTPError(
                http.StatusBadRequest,
                "User ID is missing",
            )
        }
        if err == Error.InvalidValue {
            return "", echo.NewHTTPError(
                http.StatusBadRequest,
                "User ID has an invalid format (expected: UUID)",
            )
        }
    }

    return uid, nil
}

func newUserFilter(ctx echo.Context, uid string) (*UserDTO.Filter, error) {
    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
        return nil, controller.ConvertErrorStatusToHTTP(err)
    }

    // we can trust claims if token is valid
    filter, err := UserMapper.FilterDTOFromClaims(uid, accessToken.Claims.(jwt.MapClaims))

    if err != nil {
        return nil, controller.ConvertErrorStatusToHTTP(err)
    }

    return filter, nil
}

func Create(ctx echo.Context) error {
    var body datamodel.LoginPasswordBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

    if err := DB.Database.Create(body.Login, body.Password); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if config.App.IsLoginEmail {
        activ, err := DB.Database.FindActivationByUserLogin(body.Login)
        if err != nil {
            return controller.ConvertErrorStatusToHTTP(err)
        }

        err = email.CreateAndEnqueueActivationEmail(body.Login, activ.Token)
        if err != nil {
            return controller.ConvertErrorStatusToHTTP(err)
        }
    }
    return ctx.NoContent(http.StatusOK)
}

type updater = func (*UserDTO.Filter) *Error.Status

// Updates user's state (deletion status).
// if omitUid is true, then uid will be set to empty string,
// otherwise uid will be taken from path params (in this case uid must be a valid UUID).
// If you want to change other user properties then use 'update' isntead.
func handleUserDeleteUpdate(ctx echo.Context, upd updater, omitUid bool) error {
    var uid string

    if !omitUid {
        var e *echo.HTTPError

        uid, e = getUserIdFromPath(ctx)
        if e != nil {
            return e
        }
    }

    filter, err := newUserFilter(ctx, uid)
    if err != nil {
        return err
    }

    if err := upd(filter); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    return ctx.NoContent(http.StatusOK)
}

func SoftDelete(ctx echo.Context) error {
    return handleUserDeleteUpdate(ctx, DB.Database.SoftDelete, false)
}

func Restore(ctx echo.Context) error {
    return handleUserDeleteUpdate(ctx, DB.Database.Restore, false)
}

func Drop(ctx echo.Context) error {
    return handleUserDeleteUpdate(ctx, DB.Database.Drop, false)
}

func DropAllDeleted(ctx echo.Context) error {
    return handleUserDeleteUpdate(ctx, DB.Database.DropAllSoftDeleted, true)
}

func validateSelfUpdate(filter *UserDTO.Filter, body datamodel.UpdateUserRequestBody) *echo.HTTPError {
    if filter.RequesterUID == filter.TargetUID {
        if err := body.Validate(); err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, err.Error())
        }

        user, err := DB.Database.FindAnyUserByID(filter.TargetUID)

        if err != nil {
            return controller.ConvertErrorStatusToHTTP(err)
        }

        if err := authentication.CompareHashAndPassword(user.Password, body.GetPassword()); err != nil {
            return echo.NewHTTPError(err.Status(), "Неверный пароль")
        }

        return nil
    }

    if err := body.Validate(); err != nil {
        switch body.(type){
        case *datamodel.ChangePasswordBody:
            if err == datamodel.MissingNewPassword || err == datamodel.InvalidNewPassword {
                return nil
            }
        default:
            if err == datamodel.MissingPassword || err == datamodel.InvalidPassword {
                return nil
            }
        }
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    return nil
}

// TODO try to find a way to merge 'update' and 'handleUserStateUpdate'

// Updates one of user's properties excluding state (deletion status).
// If you want to update user's state use 'handleUserStateUpdate' instead.
func update(ctx echo.Context, body datamodel.UpdateUserRequestBody) error {
    if err := ctx.Bind(body); err != nil {
        return err
    }

    uid, er := getUserIdFromPath(ctx)
    if er != nil {
        return er
    }

    filter, err := newUserFilter(ctx, uid)
    if err != nil {
        return err
    }

    if err := validateSelfUpdate(filter, body); err != nil {
        return err
    }

    var e *Error.Status

    switch b := body.(type) {
    case *datamodel.ChangeLoginBody:
        e = DB.Database.ChangeLogin(filter, b.Login)
    case *datamodel.ChangePasswordBody:
        e = DB.Database.ChangePassword(filter, b.NewPassword)
    case *datamodel.ChangeRolesBody:
        e = DB.Database.ChangeRoles(filter, b.Roles)
    default:
        return errors.New("Invalid update call: received unacceptable request body")
    }

    if e != nil {
        return controller.ConvertErrorStatusToHTTP(e)
    }

    return ctx.NoContent(http.StatusOK)
}

func ChangeLogin(ctx echo.Context) error {
    return update(ctx, new(datamodel.ChangeLoginBody))
}

func ChangePassword(ctx echo.Context) error {
    return update(ctx, new(datamodel.ChangePasswordBody))
}

func ChangeRoles(ctx echo.Context) error {
    return update(ctx, new(datamodel.ChangeRolesBody))
}

func GetRoles(ctx echo.Context) error {
    uid, er := getUserIdFromPath(ctx)
    if er != nil {
        return er
    }

    filter, err := newUserFilter(ctx, uid)
    if err != nil {
        return err
    }

    roles, e := DB.Database.GetRoles(filter)
    if e != nil {
        return controller.ConvertErrorStatusToHTTP(e)
    }

    return ctx.JSON(
        http.StatusOK,
        datamodel.RolesResponseBody{ Roles: roles },
    )
}

func IsLoginAvailable(ctx echo.Context) error {
    login := ctx.QueryParam("login")

    if login == "" {
        return echo.NewHTTPError(
            http.StatusBadRequest,
            "query param 'login' isn't specified",
        )
    }

    if err := user.VerifyLogin(login); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    exists, e := DB.Database.IsLoginAvailable(login)

    if e != nil {
        return controller.ConvertErrorStatusToHTTP(e)
    }

    return ctx.JSON(
        http.StatusOK,
        datamodel.LoginExistanceResponseBody{
            Exists: exists,
        },
    )
}

