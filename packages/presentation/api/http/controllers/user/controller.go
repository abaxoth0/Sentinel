package usercontroller

import (
	"errors"
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authentication"
	UserMapper "sentinel/packages/infrastructure/mappers"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	datamodel "sentinel/packages/presentation/data"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func newUserFilter(ctx echo.Context, uid string) (*UserDTO.Filter, error) {
    authHeader := ctx.Request().Header.Get("Authorization")
    accessToken, err := token.GetAccessToken(authHeader)

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

    return ctx.NoContent(http.StatusOK)
}

type stateUpdater = func (*UserDTO.Filter) *Error.Status

// Updates user's state (deletion status).
// If you want to change other user properties then use 'update' isntead.
func handleUserStateUpdate(ctx echo.Context, upd stateUpdater) error {
    var body datamodel.UidBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

    filter, err := newUserFilter(ctx, body.UID)

    if err != nil {
        return err
    }

    if err := upd(filter); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
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

        if err := authentication.CompareHashAndPassword(filter.TargetUID, password); err != nil {
            return echo.NewHTTPError(err.Status(), "Неверный пароль")
        }
    }

    return nil
}

// TODO try to find a way to merge 'update' and 'handleUserStateUpdate'

// Updates one of user's properties excluding state (deletion status).
// If you want to update user's state use 'handleUserStateUpdate' instead.
func update(ctx echo.Context, body datamodel.UpdateUserRequestBody) error {
    if err := controller.BindAndValidate(ctx, body); err != nil {
        return err
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
    uid := ctx.Param("uid")

    if err := validation.UUID(uid); err != nil {
        if err == Error.NoValue {
            return echo.NewHTTPError(
                http.StatusBadRequest,
                "User ID is missing",
            )
        }
        if err == Error.InvalidValue {
            return echo.NewHTTPError(
                http.StatusBadRequest,
                "The user ID has an invalid format (expected: UUID)",
            )
        }
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

func IsLoginExists(ctx echo.Context) error {
    var body datamodel.LoginBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

    exists, e := DB.Database.IsLoginExists(body.Login)

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

