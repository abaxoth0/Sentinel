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
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	RequestBody "sentinel/packages/presentation/data/request"
	ResponseBody "sentinel/packages/presentation/data/response"
	"strconv"

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
// @Router			/v1/user [post]
func Create(ctx echo.Context) error {
	var body RequestBody.LoginAndPassword

    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

    reqMeta := request.GetMetadata(ctx)

    uid, err := DB.Database.Create(body.Login, body.Password)
    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if config.App.IsLoginEmail {
        if err = email.CreateAndEnqueueActivationEmail(uid, body.Login); err != nil {
            return controller.ConvertErrorStatusToHTTP(err)
        }
    }

    controller.Log.Info("Creating new user: OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

type updater = func (*ActionDTO.UserTargeted) *Error.Status

// Updates user's state (deletion status).
// if omitUid is true, then uid will be set to empty string,
// otherwise uid will be taken from path params (in this case uid must be a valid UUID).
// If you want to change other user properties then use 'update' isntead.
func handleUserStateUpdate(ctx echo.Context, upd updater, omitUid bool, logMessageBase string) error {
    reqMeta := request.GetMetadata(ctx)

    controller.Log.Info(logMessageBase + "...", reqMeta)

    var uid string

    if !omitUid {
        uid = ctx.Param("uid")
    }

	var body RequestBody.ActionReason

	controller.Log.Info("Binding request...", reqMeta)

	if err := ctx.Bind(&body); err != nil {
		// Action reason is optional, so even if binding failed this won't be a critical problem
		controller.Log.Error("Failed to bind request", err.Error(), reqMeta)
	} else {
		controller.Log.Info("Binding request: OK", reqMeta)
	}

    act := controller.GetBasicAction(ctx).ToUserTargeted(uid)

	act.Reason = body.Reason

    if err := upd(act); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Log.Info(logMessageBase + ": OK", reqMeta)

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
// @Router			/v1/user/{uid} [delete]
// @Security		BearerAuth
// @Security		CSRF_Header
// @Security		CSRF_Cookie
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
// @Router			/v1/user/{uid}/restore [put]
// @Security		BearerAuth
// @Security		CSRF_Header
// @Security		CSRF_Cookie
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
// @Router			/v1/user/{uid}/drop [delete]
// @Security		BearerAuth
// @Security		CSRF_Header
// @Security		CSRF_Cookie
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
// @Router			/v1/user [delete]
// @Security		BearerAuth
// @Security		CSRF_Header
// @Security		CSRF_Cookie
func BulkSoftDelete(ctx echo.Context) error {
    act := controller.GetBasicAction(ctx)

	var body RequestBody.UsersIDs

	if e := controller.BindAndValidate(ctx, &body); e != nil {
		return e
	}

	act.Reason = body.Reason

    if err := DB.Database.BulkSoftDelete(act, body.IDs); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

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
// @Router			/v1/user [put]
// @Security		BearerAuth
// @Security		CSRF_Header
// @Security		CSRF_Cookie
func BulkRestore(ctx echo.Context) error {
    act := controller.GetBasicAction(ctx)

	var body RequestBody.UsersIDs

	if e := controller.BindAndValidate(ctx, &body); e != nil {
		return e
	}

	act.Reason = body.Reason

    if err := DB.Database.BulkRestore(act, body.IDs); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

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
// @Router			/v1/user/drop/all [delete]
// @Security		BearerAuth
// @Security		CSRF_Header
// @Security		CSRF_Cookie
func DropAllDeleted(ctx echo.Context) error {
    act := controller.GetBasicAction(ctx)

    if err := DB.Database.DropAllSoftDeleted(act); err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

    return ctx.NoContent(http.StatusOK)
}

func validateUpdateRequestBody(filter *ActionDTO.UserTargeted, body RequestBody.UpdateUser) *echo.HTTPError {
    // if user tries to update himself
    if filter.RequesterUID == filter.TargetUID {
        if err := body.Validate(); err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, err.Error())
        }

        user, err := DB.Database.GetAnyUserByID(filter.TargetUID)

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

    controller.Log.Info(logMessageBase + "...", reqMeta)

    controller.Log.Trace("Binding request...", reqMeta)

    if err := ctx.Bind(body); err != nil {
        controller.Log.Error("Failed to bind request", err.Error(), reqMeta)
        return err
    }

    controller.Log.Trace("Binding request: OK", reqMeta)

    uid := ctx.Param("uid")

    act := controller.GetBasicAction(ctx).ToUserTargeted(uid)

    controller.Log.Trace("Validating user update request...", reqMeta)

    if e := validateUpdateRequestBody(act, body); e != nil {
        controller.Log.Error("Invalid user update request", e.Error(), reqMeta)
        return e
    }

    controller.Log.Trace("Validating user update request: OK", reqMeta)

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
		controller.Log.Panic(
			"Invalid update call",
			fmt.Sprintf("Unexpected request body type - %T", body),
			reqMeta,
		)
        return nil
    }

    if err != nil {
		controller.Log.Info(logMessageBase + ": FAILED", reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Log.Info(logMessageBase + ": OK", reqMeta)

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
// @Router			/v1/user/{uid}/login [patch]
// @Security		BearerAuth
// @Security		CSRF_Header
// @Security		CSRF_Cookie
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
// @Router			/v1/user/{uid}/password [patch]
// @Security		BearerAuth
// @Security		CSRF_Header
// @Security		CSRF_Cookie
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
// @Router			/v1/user/{uid}/roles [patch]
// @Security		BearerAuth
// @Security		CSRF_Header
// @Security		CSRF_Cookie
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
// @Router			/v1/user/{uid}/roles [get]
// @Security		BearerAuth
func GetRoles(ctx echo.Context) error {
    uid := ctx.Param("uid")

    filter := controller.GetBasicAction(ctx).ToUserTargeted(uid)

    roles, err := DB.Database.GetRoles(filter)
    if err != nil {
        return controller.ConvertErrorStatusToHTTP(err)
    }

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
// @Router			/v1/user/login/available [get]
func IsLoginAvailable(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

	login := ctx.QueryParam("login")

    if login == "" {
		message := "query param 'login' isn't specified"

		controller.Log.Error("Failed to check if login '"+login+"' available", message, reqMeta)

        return echo.NewHTTPError(http.StatusBadRequest, message)
    }

    available := DB.Database.IsLoginInUse(login)

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
// @Router			/v1/user/search [get]
// @Security		BearerAuth
func SearchUsers(ctx echo.Context) error {
    reqMeta := request.GetMetadata(ctx)

	rawFilters := ctx.QueryParams()["filter"]
	rawPage := ctx.QueryParam("page")
	rawPageSize := ctx.QueryParam("pageSize")

	if rawFilters == nil || len(rawFilters) == 0 {
		errMsg := "Filter is missing"
		controller.Log.Error("Failed to search users", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusBadRequest, errMsg)
	}
	if rawPage == "" {
		errMsg := "Query param 'page' is missing"
		controller.Log.Error("Failed to search users", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusBadRequest, errMsg)
	}
	if rawPageSize == "" {
		errMsg := "Query param 'pageSize' is missing"
		controller.Log.Error("Failed to search users", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusBadRequest, errMsg)
	}

	page, parseErr := strconv.Atoi(rawPage)
	if parseErr != nil {
		errMsg := "page must be an integer number"
		controller.Log.Error("Failed to search users", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusBadRequest, errMsg)
	}
	pageSize, parseErr := strconv.Atoi(rawPageSize)
	if parseErr != nil {
		errMsg := "pageSize must be an integer number"
		controller.Log.Error("Failed to search users", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusBadRequest, errMsg)
	}

	act := controller.GetBasicAction(ctx)

	dtos, err := DB.Database.SearchUsers(act, rawFilters, page, pageSize)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

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
// @Router			/v1/user/{uid}/sessions [get]
// @Security		BearerAuth
func GetUserSessions(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	uid := ctx.Param("uid")

	if e := validation.UUID(uid); e != nil {
		errMsg := e.ToStatus(
			"User ID is missing in URL path",
			"User ID has invalid format (expected UUID)",
		).Error()
		controller.Log.Error("Failed to get user sessions", errMsg, reqMeta)
		return echo.NewHTTPError(http.StatusBadRequest, errMsg)
	}

	payload := controller.GetUserPayload(ctx)

	act := ActionDTO.NewUserTargeted(uid, payload.ID, payload.Roles)

	// Get locations for this sessions (in SQL query?)
	sessions, err := DB.Database.GetUserSessions(act)
	if err != nil {
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

	return ctx.JSON(http.StatusOK, res)
}

// @Summary 		Users search
// @Description 	Search users with pagination
// @ID 				get-user-by-id
// @Tags			user
// @Param 			uid 		path 	string 	true 	"User ID"
// @Accept			json
// @Produce			json
// @Success			200				{object}	userdto.Full
// @Failure			400,401,403,500	{object} 	responsebody.Error
// @Failure			490 			{object} 	responsebody.Error 			"User data desynchronization"
// @Header 			490 			{string} 	X-Token-Refresh-Required 	"Set to 'true' when token refresh is required"
// @Failure			491 			{object} 	responsebody.Error 			"Session revoked"
// @Header 			491 			{string} 	X-Session-Revoked 			"Set to 'true' if current user session was revoked"
// @Router			/v1/user/{uid} [get]
// @Security		BearerAuth
func GetUser(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	uid := ctx.Param("uid")

	if e := validation.UUID(uid); e != nil {
		errMSg := e.ToStatus(
			"User ID is missing in URL path",
			"User ID has invalid format (expected UUID)",
		).Error()
		controller.Log.Error("Failed to get user", errMSg, reqMeta)
		return echo.NewHTTPError(http.StatusBadRequest, errMSg)
	}

	payload := controller.GetUserPayload(ctx)

	act := ActionDTO.NewUserTargeted(uid, payload.ID, payload.Roles)

	err := authz.User.GetUserSession(act.RequesterUID == act.TargetUID, act.RequesterRoles)
	if err != nil {
		return err
	}

	user, err := DB.Database.GetUserByID(uid)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, user)
}

