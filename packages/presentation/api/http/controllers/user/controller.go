package usercontroller

import (
	"errors"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authentication"
	UserMapper "sentinel/packages/infrastructure/mappers"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/response"
	datamodel "sentinel/packages/presentation/data"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func newUserFilter(ctx echo.Context, uid string) (*UserDTO.Filter, error) {
    authHeader := ctx.Request().Header.Get("Authorization")
    accessToken, err := token.GetAccessToken(authHeader)

    if err != nil {
        return nil, echo.NewHTTPError(err.Status, err.Message)
    }

    // we can trust claims if token is valid
    filter, err := UserMapper.FilterDTOFromClaims(uid, accessToken.Claims.(jwt.MapClaims))

    if err != nil {
        return nil, echo.NewHTTPError(err.Status, err.Message)
    }

    return filter, nil
}

func Create(ctx echo.Context) error {
    var body datamodel.LoginPasswordBody

    if err := ctx.Bind(&body); err != nil {
        return err
    }

    if err := body.Validate(); err != nil {
        return echo.NewHTTPError(
            http.StatusBadRequest,
            err.Error(),
            )
    }

    err := DB.Database.Create(body.Login, body.Password)

    if err != nil {
        if is, e := Error.IsStatusError(err); is {
            return ctx.JSON(
                e.Status,
                datamodel.MessageResponseBody{
                    Message: e.Message,
                },
            )
        }

        return err
    }

    return ctx.NoContent(http.StatusOK)
}

type stateUpdater = func (*UserDTO.Filter) *Error.Status

// Updates user's state (deletion status).
// If you want to change other user properties then use 'update' isntead.
func handleUserStateUpdate(ctx echo.Context, upd stateUpdater) error {
    var body datamodel.UidBody

    if err := ctx.Bind(&body); err != nil {
        return err
    }

    if err := body.Validate(); err != nil {
        return response.RequestMissingUid
    }

    filter, err := newUserFilter(ctx, body.UID)

    if err != nil {
        return err
    }

    if err := upd(filter); err != nil {
        return echo.NewHTTPError(err.Status, err.Message)
    }

    return ctx.NoContent(http.StatusOK)
}

func SoftDelete(ctx echo.Context) error {
    return handleUserStateUpdate(ctx, DB.Database.SoftDelete)
}

func Restore(ctx echo.Context) error {
    return handleUserStateUpdate(ctx, DB.Database.Restore)
}

func Drop(ctx echo.Context) error {
    return handleUserStateUpdate(ctx, DB.Database.Drop)
}

func DropAllDeleted(ctx echo.Context) error {
    return handleUserStateUpdate(ctx, DB.Database.DropAllSoftDeleted)
}

func validateSelfUpdate(filter *UserDTO.Filter, password string) *echo.HTTPError {
    if filter.RequesterUID == filter.TargetUID {
        if password == "" {
            return echo.NewHTTPError(
                http.StatusUnprocessableEntity,
                "Password required when modifying your own account",
            )
        }

        err := authentication.ComparePasswords(filter.TargetUID, password)

        if err != nil {
            return echo.NewHTTPError(err.Status, err.Message)
        }
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

    if err := body.Validate(); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    filter, err := newUserFilter(ctx, body.GetUID())

    if err != nil {
        return err
    }

    if err := validateSelfUpdate(filter, body.GetPassword()); err != nil {
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
        return echo.NewHTTPError(e.Status, e.Message)
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
    var body datamodel.UidBody

    if err := ctx.Bind(&body); err != nil {
        return err
    }

    filter, err := newUserFilter(ctx, body.UID)

    if err != nil {
        return err
    }

    roles, e := DB.Database.GetRoles(filter)

    if e != nil {
        return echo.NewHTTPError(e.Status, e.Message)
    }

    return ctx.JSON(
        http.StatusOK,
        datamodel.RolesResponseBody{
            Roles: roles,
        },
    )
}

func IsLoginExists(ctx echo.Context) error {
    var body datamodel.LoginBody

    if err := ctx.Bind(&body); err != nil {
        return err
    }

    if err := body.Validate(); err != nil {
        return response.RequestMissingLogin
    }

    exists, e := DB.Database.IsLoginExists(body.Login)

    if e != nil {
        return echo.NewHTTPError(e.Status, e.Message)
    }

    return ctx.JSON(
        http.StatusOK,
        datamodel.LoginExistanceResponseBody{
            Exists: exists,
        },
    )
}

