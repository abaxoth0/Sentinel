package usercontroller

import (
	"fmt"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authn"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/email"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	RequestBody "sentinel/packages/presentation/data/request"
	ResponseBody "sentinel/packages/presentation/data/response"
	"strconv"
	"strings"

	rbac "github.com/StepanAnanin/SentinelRBAC"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// @Summary 		Create new user
// @Description 	Registration endpoint
// @ID 				create-new-user
// @Tags			user
// @Param 			credentials body requestbody.LoginAndPassword true "User credentials"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,500 	{object} 	responsebody.Error
// @Router			/user [post]
func Create(ctx echo.Context) error {
	var body RequestBody.LoginAndPassword

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

type updater = func (*ActionDTO.UserTargeted) *Error.Status

// Updates user's state (deletion status).
// if omitUid is true, then uid will be set to empty string,
// otherwise uid will be taken from path params (in this case uid must be a valid UUID).
// If you want to change other user properties then use 'update' isntead.
func handleUserStateUpdate(ctx echo.Context, upd updater, omitUid bool, logMessageBase string) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info(logMessageBase + "...", reqMeta)

    var uid string

    if !omitUid {
        uid = ctx.Param("uid")
    }

	var body RequestBody.ActionReason

	controller.Logger.Info("Binding request...", reqMeta)

	if err := ctx.Bind(&body); err != nil {
		controller.Logger.Error("Failed to bind request", err.Error(), reqMeta)
	} else {
		controller.Logger.Info("Binding request: OK", reqMeta)
	}

    act, err := controller.NewTargetedActionDTO(ctx, uid)
    if err != nil {
		controller.Logger.Error(logMessageBase + ": FAILED", err.Error(), reqMeta)
        return err
    }

	act.Reason = body.Reason

    if err := upd(act); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Info(logMessageBase + ": OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

// @Summary 		Soft delete user
// @Description 	Soft delete user. All sessions of soft deleted user will be revoked
// @ID 				soft-delete-user
// @Tags			user
// @Param 			uid path string true "User ID"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/{uid} [delete]
// @Security		BearerAuth
func SoftDelete(ctx echo.Context) error {
	return handleUserStateUpdate(ctx, DB.Database.SoftDelete, false, "Soft deleting user")
}

// @Summary 		Restore soft delete user
// @Description 	Restore soft delete user
// @ID 				restore-soft-delete-user
// @Tags			user
// @Param 			uid path string true "User ID"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/{uid}/restore [put]
// @Security		BearerAuth
func Restore(ctx echo.Context) error {
    return handleUserStateUpdate(ctx, DB.Database.Restore, false, "Restoring user")
}

// @Summary 		Hard delete user
// @Description 	Hard delete user. Only soft deleted users can be hard deleted
// @ID 				hard-delete-user
// @Tags			user
// @Param 			uid path string true "User ID"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/{uid}/drop [delete]
// @Security		BearerAuth
func Drop(ctx echo.Context) error {
    return handleUserStateUpdate(ctx, DB.Database.Drop, false, "Dropping user")
}

// @Summary 		Soft delete several users
// @Description 	Bulk user soft delete. All sessions of soft deleted users will be revoked
// @ID 				bulk-soft-delete-users
// @Tags			user
// @Param 			usersIDs body requestbody.UsersIDs true "Users IDs"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user [delete]
// @Security		BearerAuth
func BulkSoftDelete(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Bulk soft deleting users...", reqMeta)

    act, err := controller.NewBasicActionDTO(ctx)
    if err != nil {
		controller.Logger.Error("Bulk soft deleting users: FAILED", err.Error(), reqMeta)
        return err
    }

	var body RequestBody.UsersIDs

	if e := controller.BindAndValidate(ctx, &body); e != nil {
		return e
	}

	body.Reason = act.Reason

    if err := DB.Database.BulkSoftDelete(act, body.IDs); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Info("Bulk soft deleting users: OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

// @Summary 		Restore several soft deleted users
// @Description 	Bulk restore soft deleted users
// @ID 				bulk-restore-users
// @Tags			user
// @Param 			usersIDs body requestbody.UsersIDs true "Users IDs"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user [put]
// @Security		BearerAuth
func BulkRestore(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Bulk restoring users...", reqMeta)

    act, err := controller.NewBasicActionDTO(ctx)
    if err != nil {
		controller.Logger.Error("Bulk restoring users: FAILED", err.Error(), reqMeta)
        return err
    }

	var body RequestBody.UsersIDs

	if e := controller.BindAndValidate(ctx, &body); e != nil {
		return e
	}

	act.Reason = body.Reason

    if err := DB.Database.BulkRestore(act, body.IDs); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Info("Bulk restoring users: OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

// @Summary 		Drop all delete users
// @Description 	Hard delete all soft deleted users
// @ID 				drop-all-deleted-users
// @Tags			user
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/drop/all [delete]
// @Security		BearerAuth
func DropAllDeleted(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info("Dropping all soft deleted user...", reqMeta)

    act, err := controller.NewTargetedActionDTO(ctx, "")
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

func validateUpdateRequestBody(filter *ActionDTO.UserTargeted, body RequestBody.UpdateUser) *echo.HTTPError {
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
        if _, ok := body.(*RequestBody.ChangePassword); ok {
            if err == RequestBody.ErrorMissingPassword || err == RequestBody.ErrorInvalidPassword {
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
func update(ctx echo.Context, body RequestBody.UpdateUser, logMessageBase string) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Logger.Info(logMessageBase + "...", reqMeta)

    controller.Logger.Trace("Binding request...", reqMeta)

    if err := ctx.Bind(body); err != nil {
        controller.Logger.Error("Failed to bind request", err.Error(), reqMeta)
        return err
    }

    controller.Logger.Trace("Binding request: OK", reqMeta)

    uid := ctx.Param("uid")

    act, e := controller.NewTargetedActionDTO(ctx, uid)
    if e != nil {
        return e
    }

    controller.Logger.Trace("Validating user update request...", reqMeta)

    if e := validateUpdateRequestBody(act, body); e != nil {
        controller.Logger.Error("Invalid user update request", e.Error(), reqMeta)
        return e
    }

    controller.Logger.Trace("Validating user update request: OK", reqMeta)

	if act.TargetUID != act.RequesterUID {
		act.Reason = body.GetReason()
	}

    var err *Error.Status

    switch b := body.(type) {
    case *RequestBody.ChangeLogin:
        err = DB.Database.ChangeLogin(act, b.Login)
    case *RequestBody.ChangePassword:
        err = DB.Database.ChangePassword(act, b.NewPassword)
    case *RequestBody.ChangeRoles:
        err = DB.Database.ChangeRoles(act, b.Roles)
    default:
		controller.Logger.Panic(
			"Invalid update call",
			fmt.Sprintf("Unexpected request body type - %T", body),
			reqMeta,
		)
        return nil
    }

    if err != nil {
		controller.Logger.Info(logMessageBase + ": FAILED", reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Info(logMessageBase + ": OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

// @Summary 		Change user login
// @Description 	Change user login
// @ID 				change-user-login
// @Tags			user
// @Param 			uid 					path 	string 								true 	"User ID"
// @Param 			newLogin 				body 	requestbody.UserLogin	 			true 	"New user login"
// @Param 			newLoginAndPassword 	body 	requestbody.LoginAndPassword		false 	"New user login and password (required if user tries to change his own login)"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/{uid}/login [patch]
// @Security		BearerAuth
func ChangeLogin(ctx echo.Context) error {
    return update(ctx, new(RequestBody.ChangeLogin), "Changing user login")
}

// @Summary 		Change user password
// @Description 	Change user password
// @ID 				change-user-password
// @Tags			user
// @Param 			uid 					path 	string 							true 	"User ID"
// @Param 			newPassword 			body 	requestbody.UserPassword 		true 	"New user password"
// @Param 			newAndCurrentPassword 	body 	requestbody.ChangePassword		false 	"Both new and current user passwords (required if user tries to change his own login)"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/{uid}/password [patch]
// @Security		BearerAuth
func ChangePassword(ctx echo.Context) error {
    return update(ctx, new(RequestBody.ChangePassword), "Changing user password")
}

// @Summary 		Change user roles
// @Description 	Change user roles
// @ID 				change-user-roles
// @Tags			user
// @Param 			uid 					path 	string 							true 	"User ID"
// @Param 			newRoles 				body 	requestbody.UserRoles	 		true 	"New user roles"
// @Param 			newRolesAndPassword 	body 	requestbody.ChangeRoles			false 	"New user roles and password (required if user tries to change his own login)"
// @Accept			json
// @Produce			json
// @Success			200
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/{uid}/roles [patch]
// @Security		BearerAuth
func ChangeRoles(ctx echo.Context) error {
    return update(ctx, new(RequestBody.ChangeRoles), "Changing user roles")
}

// @Summary 		Get user roles
// @Description 	Get user roles
// @ID 				get-user-roles
// @Tags			user
// @Param 			uid	path string true "User ID"
// @Accept			json
// @Produce			json
// @Success			200				{array}		string
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/{uid}/roles [get]
// @Security		BearerAuth
func GetRoles(ctx echo.Context) error {
    uid := ctx.Param("uid")

    filter, e := controller.NewTargetedActionDTO(ctx, uid)
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

    return ctx.JSON(http.StatusOK, roles)
}

// @Summary 		Check login availability
// @Description 	Check is login free to use
// @ID 				check-login
// @Tags			user
// @Param 			login query string true "The login you want to check"
// @Accept			json
// @Produce			json
// @Success			200				{object}	responsebody.IsLoginAvailable
// @Failure			400,401,500		{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/login/available [get]
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
        ResponseBody.IsLoginAvailable{
            Available: available,
        },
    )
}

// @Summary 		Users search
// @Description 	Search users with pagination
// @ID 				search-users
// @Tags			user
// @Param 			filter 		query 	string 	true 	"Search filter"
// @Param 			page 		query 	int 	true 	"Search page"
// @Param 			pageSize 	query 	int 	true 	"Elements per page"
// @Accept			json
// @Produce			json
// @Success			200				{object}	[]userdto.Public
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/search [get]
// @Security		BearerAuth
func SearchUsers(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

	rawFilters := ctx.QueryParams()["filter"]
	rawPage := ctx.QueryParam("page")
	rawPageSize := ctx.QueryParam("pageSize")

	if rawFilters == nil || len(rawFilters) == 0 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Filter is missing",
		)
	}
	if rawPage == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Query param 'page' is missing")
	}
	if rawPageSize == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Query param 'pageSize' is missing")
	}

	page, parseErr := strconv.Atoi(rawPage)
	if parseErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "page must be an integer number")
	}
	pageSize, parseErr := strconv.Atoi(rawPageSize)
	if parseErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "pageSize must be an integer number")
	}

	act, e := controller.NewBasicActionDTO(ctx)
	if e != nil {
		return e
	}

	controller.Logger.Info("Searching for users matching '"+strings.Join(rawFilters, ";")+"' filters...", reqMeta)

	dtos, err := DB.Database.SearchUsers(act, rawFilters, page, pageSize)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	controller.Logger.Info("Searching for users matching '"+strings.Join(rawFilters, ";")+"' filters: OK", reqMeta)

	return ctx.JSON(http.StatusOK, dtos)
}

// @Summary 		Get user sessions
// @Description 	Get all active user sessions
// @ID 				get-user-sessions
// @Tags			user
// @Param 			uid path string true "User ID"
// @Accept			json
// @Produce			json
// @Success			200				{object}	[]responsebody.UserSession
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/user/{uid}/sessions [get]
// @Security		BearerAuth
func GetUserSessions(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	uid := ctx.Param("uid")

	if e := validation.UUID(uid); e != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			e.ToStatus(
				"User ID is missing in URL path",
				"User ID has invalid format (expected UUID)",
			).Error(),
		)
	}

	accessToken, err := controller.GetAccessToken(ctx)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	payload, err := UserMapper.PayloadFromClaims(accessToken.Claims.(jwt.MapClaims))
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	controller.Logger.Info("Getting user sessions...", reqMeta)

	act := ActionDTO.NewUserTargeted(uid, payload.ID, payload.Roles)

	// Get locations for this sessions (in SQL query?)
	sessions, err := DB.Database.GetUserSessions(act)
	if err != nil {
		controller.Logger.Error("Failed to get user sessions", err.Error(), reqMeta)
		return controller.ConvertErrorStatusToHTTP(err)
	}

	res := make([]ResponseBody.UserSession, 0, len(sessions))

	for _, session := range sessions {
		location, err := DB.Database.GetLocationBySessionID(act, session.ID)
		if err == nil {
			res = append(res, ResponseBody.UserSession{
				Session: session,
				Location: location.MakePublic(),
			})
		}
	}

	controller.Logger.Info("Getting user sessions: OK", reqMeta)

	return ctx.JSON(http.StatusOK, res)
}

