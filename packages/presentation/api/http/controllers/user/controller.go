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
	"sentinel/packages/presentation/api/http/request"
	datamodel "sentinel/packages/presentation/data"
	"strconv"

	rbac "github.com/StepanAnanin/SentinelRBAC"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func newTargetedActionDTO(ctx echo.Context, uid string) (*ActionDTO.Targeted, error) {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Trace("Retrieving access token from the request...", reqMeta)

    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
        controller.Logger.Error("Failed to retrieve valid access token from the request", err.Error(), reqMeta)
        return nil, controller.HandleTokenError(ctx, err)
    }

    controller.Logger.Trace("Retrieving access token from the request: OK", reqMeta)
    controller.Logger.Trace("Creating action DTO from token claims...", reqMeta)

    // claims can be trusted if token is valid
    act, err := UserMapper.TargetedActionDTOFromClaims(uid, accessToken.Claims.(jwt.MapClaims))
    if err != nil {
        controller.Logger.Error("Failed to create action DTO from token claims", err.Error(), reqMeta)
        return nil, controller.ConvertErrorStatusToHTTP(err)
    }

    controller.Logger.Trace("Creating action DTO from token claims: OK", reqMeta)

    return act, nil
}

func Create(ctx echo.Context) error {
    var body datamodel.LoginPasswordBody

    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Creating new user...", reqMeta)

    uid, err := DB.Database.Create(body.Login, body.Password)
    if err != nil {
        controller.Logger.Error("Failed to create new user", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if config.App.IsLoginEmail {
        controller.Logger.Trace("Creating activation token...", reqMeta)

        tk, err := token.NewActivationToken(
            uid,
            body.Login,
            rbac.GetRolesNames(authz.Host.DefaultRoles),
        )
        if err != nil {
            controller.Logger.Error("Failed to create new activation token", err.Error(), reqMeta)
            return controller.ConvertErrorStatusToHTTP(err)
        }

        controller.Logger.Trace("Creating activation token: OK", reqMeta)
        controller.Logger.Trace("Creating and equeueing activation email...", reqMeta)

        err = email.CreateAndEnqueueActivationEmail(body.Login, tk.String())
        if err != nil {
            controller.Logger.Error("Failed to create and enqueue activation email", err.Error(), reqMeta)
            return controller.ConvertErrorStatusToHTTP(err)
        }

        controller.Logger.Trace("Creating and equeueing activation email: OK", reqMeta)
    }

    controller.Logger.Info("Creating new user: OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

type updater = func (*ActionDTO.Targeted) *Error.Status

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

// TODO a lot of boilerplate, try to do smth with that

func SoftDelete(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Soft deleting user...", reqMeta)

    if err := handleUserDeleteUpdate(ctx, DB.Database.SoftDelete, false); err != nil {
        controller.Logger.Error("Failed to soft delete user", err.Error(), reqMeta)
        return err
    }

    controller.Logger.Info("Soft deleting user: OK", reqMeta)

    return nil
}

func Restore(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Restoring user...", reqMeta)

    if err := handleUserDeleteUpdate(ctx, DB.Database.Restore, false); err != nil {
        controller.Logger.Error("Failed to restore user", err.Error(), reqMeta)
        return err
    }

    controller.Logger.Info("Restoring user: OK", reqMeta)

    return nil
}

func Drop(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Dropping user...", reqMeta)

    if err := handleUserDeleteUpdate(ctx, DB.Database.Drop, false); err != nil {
        controller.Logger.Error("Failed to drop user", err.Error(), reqMeta)
        return err
    }

    controller.Logger.Info("Dropping user: OK", reqMeta)

    return nil
}

func DropAllDeleted(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Dropping all soft deleted user...", reqMeta)

    act, err := newTargetedActionDTO(ctx, "")
    if err != nil {
        return err
    }

    if err := DB.Database.DropAllSoftDeleted(&act.Basic); err != nil {
        controller.Logger.Error("Failed to drop all soft deleted user", err.Error(), reqMeta)

        return controller.ConvertErrorStatusToHTTP(err)
    }

    controller.Logger.Info("Dropping all soft deleted user: Ok", reqMeta)

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
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Trace("Binding request...", reqMeta)

    if err := ctx.Bind(body); err != nil {
        controller.Logger.Error("Failed to bind request", err.Error(), reqMeta)
        return err
    }

    controller.Logger.Trace("Binding request: OK", reqMeta)

    uid := ctx.Param("uid")

    filter, e := newTargetedActionDTO(ctx, uid)
    if e != nil {
        return e
    }

    controller.Logger.Trace("Validating user update request...", reqMeta)

    if e := validateUpdateRequestBody(filter, body); e != nil {
        controller.Logger.Error("Invalid user update request", e.Error(), reqMeta)
        return e
    }

    controller.Logger.Trace("Validating user update request: OK", reqMeta)

    var err *Error.Status

    switch b := body.(type) {
    case *datamodel.ChangeLoginBody:
        err = DB.Database.ChangeLogin(filter, b.Login)
    case *datamodel.ChangePasswordBody:
        err = DB.Database.ChangePassword(filter, b.NewPassword)
    case *datamodel.ChangeRolesBody:
        err = DB.Database.ChangeRoles(filter, b.Roles)
    default:
        return errors.New("Invalid update call: received unacceptable request body")
    }

    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    return ctx.NoContent(http.StatusOK)
}

func ChangeLogin(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Changing user login...", reqMeta)

    if err := update(ctx, new(datamodel.ChangeLoginBody)); err != nil {
        controller.Logger.Error("Failed to change user login", err.Error(), reqMeta)
        return err
    }

    controller.Logger.Info("Changing user login: OK", reqMeta)

    return nil
}

func ChangePassword(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Changing user password...", reqMeta)

    if err := update(ctx, new(datamodel.ChangePasswordBody)); err != nil {
        controller.Logger.Error("Failed to change user password", err.Error(), reqMeta)
        return err
    }

    controller.Logger.Info("Changing user password: OK", reqMeta)

    return nil
}

func ChangeRoles(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Changing user roles...", reqMeta)

    if err := update(ctx, new(datamodel.ChangeRolesBody)); err != nil {
        controller.Logger.Error("Failed to change user roles", err.Error(), reqMeta)
        return err
    }

    controller.Logger.Info("Changing user roles: OK", reqMeta)

    return nil
}

func GetRoles(ctx echo.Context) error {
    uid := ctx.Param("uid")

    filter, e := newTargetedActionDTO(ctx, uid)
    if e != nil {
        return e
    }

    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Getting user roles...", reqMeta)

    roles, err := DB.Database.GetRoles(filter)
    if err != nil {
        controller.Logger.Error("Failed to get user roles", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

    controller.Logger.Info("Getting user roles: OK", reqMeta)

    return ctx.JSON(
        http.StatusOK,
        datamodel.RolesResponseBody{ Roles: roles },
    )
}

func IsLoginAvailable(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

	login := ctx.QueryParam("login")

	controller.Logger.Info("Checking if login '"+login+"' available...", reqMeta)

    if login == "" {
		message := "query param 'login' isn't specified"

		controller.Logger.Error("Failed to check if login '"+login+"' available", message, reqMeta)

        return echo.NewHTTPError(
            http.StatusBadRequest,
			message,
        )
    }

    available := DB.Database.IsLoginAvailable(login)

	controller.Logger.Info(
		"Checking if login '"+login+"' available: " + strconv.FormatBool(available), reqMeta,
	)

    return ctx.JSON(
        http.StatusOK,
        datamodel.IsLoginAvailableResponseBody{
            Available: available,
        },
    )
}

