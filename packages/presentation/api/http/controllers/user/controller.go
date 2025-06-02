package usercontroller

import (
	"errors"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authn"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/email"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	datamodel "sentinel/packages/presentation/data"
	"strconv"

	rbac "github.com/StepanAnanin/SentinelRBAC"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func newTargetedActionDTO(ctx echo.Context, uid string) (*ActionDTO.Targeted, error) {
    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Trace("Retrieving access token from the request..." + reqInfo)

    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
        controller.Logger.Error("Failed to retrieve valid access token from the request" + reqInfo, err.Error())
        return nil, controller.HandleTokenError(ctx, err)
    }

    controller.Logger.Trace("Retrieving access token from the request: OK" + reqInfo)
    controller.Logger.Trace("Creating action DTO from token claims..." + reqInfo)

    // claims can be trusted if token is valid
    act, err := UserMapper.TargetedActionDTOFromClaims(uid, accessToken.Claims.(jwt.MapClaims))
    if err != nil {
        controller.Logger.Error("Failed to create action DTO from token claims" + reqInfo, err.Error())
        return nil, controller.ConvertErrorStatusToHTTP(err)
    }

    controller.Logger.Trace("Creating action DTO from token claims: OK" + reqInfo)

    return act, nil
}

func Create(ctx echo.Context) error {
    var body datamodel.LoginPasswordBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Info("Creating new user..." + reqInfo)

    uid, err := DB.Database.Create(body.Login, body.Password)
    if err != nil {
        controller.Logger.Error("Failed to create new user" + reqInfo, err.Error())
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if config.App.IsLoginEmail {
        controller.Logger.Trace("Creating activation token..." + reqInfo)

        tk, err := token.NewActivationToken(
            uid,
            body.Login,
            rbac.GetRolesNames(authz.Host.DefaultRoles),
        )
        if err != nil {
            controller.Logger.Error("Failed to create new activation token" + reqInfo, err.Error())
            return controller.ConvertErrorStatusToHTTP(err)
        }

        controller.Logger.Trace("Creating activation token: OK" + reqInfo)
        controller.Logger.Trace("Creating and equeueing activation email..." + reqInfo)

        err = email.CreateAndEnqueueActivationEmail(body.Login, tk.String())
        if err != nil {
            controller.Logger.Error("Failed to create and enqueue activation email" + reqInfo, err.Error())
            return controller.ConvertErrorStatusToHTTP(err)
        }

        controller.Logger.Trace("Creating and equeueing activation email: OK" + reqInfo)
    }

    controller.Logger.Info("Creating new user: OK" + reqInfo)

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
    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Info("Soft deleting user..." + reqInfo)

    if err := handleUserDeleteUpdate(ctx, DB.Database.SoftDelete, false); err != nil {
        controller.Logger.Error("Failed to soft delete user" + reqInfo, err.Error())
        return err
    }

    controller.Logger.Info("Soft deleting user: OK" + reqInfo)

    return nil
}

func Restore(ctx echo.Context) error {
    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Info("Restoring user..." + reqInfo)

    if err := handleUserDeleteUpdate(ctx, DB.Database.Restore, false); err != nil {
        controller.Logger.Error("Failed to restore user" + reqInfo, err.Error())
        return err
    }

    controller.Logger.Info("Restoring user: OK" + reqInfo)

    return nil
}

func Drop(ctx echo.Context) error {
    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Info("Dropping user..." + reqInfo)

    if err := handleUserDeleteUpdate(ctx, DB.Database.Drop, false); err != nil {
        controller.Logger.Error("Failed to drop user" + reqInfo, err.Error())
        return err
    }

    controller.Logger.Info("Dropping user: OK" + reqInfo)

    return nil
}

func DropAllDeleted(ctx echo.Context) error {
    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Info("Dropping all soft deleted user..." + reqInfo)

    act, err := newTargetedActionDTO(ctx, "")
    if err != nil {
        return err
    }

    if err := DB.Database.DropAllSoftDeleted(&act.Basic); err != nil {
        controller.Logger.Error("Failed to drop all soft deleted user" + reqInfo, err.Error())

        return controller.ConvertErrorStatusToHTTP(err)
    }

    controller.Logger.Info("Dropping all soft deleted user: Ok" + reqInfo)

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
    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Trace("Binding request..." + reqInfo)

    if err := ctx.Bind(body); err != nil {
        controller.Logger.Error("Failed to bind request" + reqInfo, err.Error())
        return err
    }

    controller.Logger.Trace("Binding request: OK" + reqInfo)

    uid := ctx.Param("uid")

    filter, err := newTargetedActionDTO(ctx, uid)
    if err != nil {
        return err
    }

    controller.Logger.Trace("Validating user update request..." + reqInfo)

    if err := validateUpdateRequestBody(filter, body); err != nil {
        controller.Logger.Error("Invalid user update request" + reqInfo, err.Error())
        return err
    }

    controller.Logger.Trace("Validating user update request: OK" + reqInfo)

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
    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Info("Changing user login..." + reqInfo)

    if err := update(ctx, new(datamodel.ChangeLoginBody)); err != nil {
        controller.Logger.Error("Failed to change user login" + reqInfo, err.Error())
        return err
    }

    controller.Logger.Info("Changing user login: OK" + reqInfo)

    return nil
}

func ChangePassword(ctx echo.Context) error {
    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Info("Changing user password..." + reqInfo)

    if err := update(ctx, new(datamodel.ChangePasswordBody)); err != nil {
        controller.Logger.Error("Failed to change user password" + reqInfo, err.Error())
        return err
    }

    controller.Logger.Info("Changing user password: OK" + reqInfo)

    return nil
}

func ChangeRoles(ctx echo.Context) error {
    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Info("Changing user roles..." + reqInfo)

    if err := update(ctx, new(datamodel.ChangeRolesBody)); err != nil {
        controller.Logger.Error("Failed to change user roles" + reqInfo, err.Error())
        return err
    }

    controller.Logger.Info("Changing user roles: OK" + reqInfo)

    return nil
}

func GetRoles(ctx echo.Context) error {
    uid := ctx.Param("uid")

    filter, err := newTargetedActionDTO(ctx, uid)
    if err != nil {
        return err
    }

    reqInfo := controller.RequestInfo(ctx)

    controller.Logger.Info("Getting user roles..." + reqInfo)

    roles, e := DB.Database.GetRoles(filter)
    if e != nil {
        controller.Logger.Error("Failed to get user roles" + reqInfo, err.Error())
        return controller.ConvertErrorStatusToHTTP(e)
    }

    controller.Logger.Info("Getting user roles: OK" + reqInfo)

    return ctx.JSON(
        http.StatusOK,
        datamodel.RolesResponseBody{ Roles: roles },
    )
}

func IsLoginAvailable(ctx echo.Context) error {
    reqInfo := controller.RequestInfo(ctx)

	login := ctx.QueryParam("login")

	controller.Logger.Info("Checking if login '"+login+"' available..." + reqInfo)

    if login == "" {
		message := "query param 'login' isn't specified"

		controller.Logger.Error("Failed to check if login '"+login+"' available" + reqInfo, message)

        return echo.NewHTTPError(
            http.StatusBadRequest,
			message,
        )
    }

    available := DB.Database.IsLoginAvailable(login)

	controller.Logger.Info(
		"Checking if login '"+login+"' available: " + strconv.FormatBool(available) + reqInfo,
	)

    return ctx.JSON(
        http.StatusOK,
        datamodel.IsLoginAvailableResponseBody{
            Available: available,
        },
    )
}

