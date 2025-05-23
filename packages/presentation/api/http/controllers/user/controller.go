package usercontroller

import (
	"errors"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authn"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/email"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	datamodel "sentinel/packages/presentation/data"

	rbac "github.com/StepanAnanin/SentinelRBAC"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func newTargetedActionDTO(ctx echo.Context, uid string) (*ActionDTO.Targeted, error) {
    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
        return nil, controller.HandleTokenError(ctx, err)
    }

    // we can trust claims if token is valid
    act, err := UserMapper.TargetedActionDTOFromClaims(uid, accessToken.Claims.(jwt.MapClaims))

    if err != nil {
        return nil, controller.ConvertErrorStatusToHTTP(err)
    }

    return act, nil
}

func Create(ctx echo.Context) error {
    var body datamodel.LoginPasswordBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

    uid, err := DB.Database.Create(body.Login, body.Password)
    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if config.App.IsLoginEmail {
        tk, err := token.NewActivationToken(
            uid,
            body.Login,
            rbac.GetRolesNames(authz.Host.DefaultRoles),
        )
        if err != nil {
            return controller.ConvertErrorStatusToHTTP(err)
        }

        err = email.CreateAndEnqueueActivationEmail(body.Login, tk.String())
        if err != nil {
            return controller.ConvertErrorStatusToHTTP(err)
        }
    }

    return ctx.NoContent(http.StatusOK)
}

type updater= func (*ActionDTO.Targeted) *Error.Status

// Updates user's state (deletion status).
// if omitUid is true, then uid will be set to empty string,
// otherwise uid will be taken from path params (in this case uid must be a valid UUID).
// If you want to change other user properties then use 'update' isntead.
func handleUserDeleteUpdate(ctx echo.Context, upd updater, omitUid bool) error {
    var uid string

    if !omitUid {
        uid = ctx.Param("uid")
    }

    act, err := newTargetedActionDTO(ctx, uid)
    if err != nil {
        return err
    }

    if err := upd(act); err != nil {
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
    act, err := newTargetedActionDTO(ctx, "")
    if err != nil {
        return err
    }

    if err := DB.Database.DropAllSoftDeleted(&act.Basic); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    return ctx.NoContent(http.StatusOK)
}

func validateUpdateRequestBody(filter *ActionDTO.Targeted, body datamodel.UpdateUserRequestBody) *echo.HTTPError {
    // if user tries to update himself
    if filter.RequesterUID == filter.TargetUID {
        if err := body.Validate(); err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, err.Error())
        }

        user, err := DB.Database.FindAnyUserByID(filter.TargetUID)

        if err != nil {
            return controller.ConvertErrorStatusToHTTP(err)
        }

        if err := authn.CompareHashAndPassword(user.Password, body.GetPassword()); err != nil {
            return echo.NewHTTPError(err.Status(), "Неверный пароль")
        }

        return nil
    }

    // if user tries to update another user
    if err := body.Validate(); err != nil {
        if _, ok := body.(*datamodel.ChangePasswordBody); ok {
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

    uid := ctx.Param("uid")

    filter, err := newTargetedActionDTO(ctx, uid)
    if err != nil {
        return err
    }

    if err := validateUpdateRequestBody(filter, body); err != nil {
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

    filter, err := newTargetedActionDTO(ctx, uid)
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

    if err := user.ValidateLogin(login); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    available, e := DB.Database.IsLoginAvailable(login)

    if e != nil {
        return controller.ConvertErrorStatusToHTTP(e)
    }

    return ctx.JSON(
        http.StatusOK,
        datamodel.IsLoginAvailableResponseBody{
            Available: available,
        },
    )
}

