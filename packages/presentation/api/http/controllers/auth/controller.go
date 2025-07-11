package authcontroller

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authn"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	RequestBody "sentinel/packages/presentation/data/request"
	ResponseBody "sentinel/packages/presentation/data/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func Login(ctx echo.Context) error {
    var body RequestBody.LoginAndPassword
    if err := controller.BindAndValidate(ctx, &body); err != nil {
        return err
    }

	reqMeta := request.GetMetadata(ctx)

	controller.Logger.Info("Authenticating user '"+body.Login+"'...", reqMeta)

    user, err := DB.Database.FindUserByLogin(body.Login)
    if err != nil {
		controller.Logger.Error("Failed to authenticate user '"+body.Login, err.Error(), reqMeta)

        if err.Side() == Error.ClientSide {
            return controller.ConvertErrorStatusToHTTP(authn.InvalidAuthCreditinals)
        }
        return controller.ConvertErrorStatusToHTTP(err)
    }

    if err := authn.CompareHashAndPassword(user.Password, body.Password); err != nil {
		controller.Logger.Error("Failed to authenticate user '"+body.Login+"'", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

	// Trying to update existing session
	if tk, err := controller.GetRefreshToken(ctx); err != nil {
		controller.Logger.Error("Failed to get refresh token from request", err.Error(), reqMeta)
	} else {
		controller.Logger.Info("Updating user session...", reqMeta)

		payload, err := UserMapper.PayloadFromClaims(tk.Claims.(jwt.MapClaims))
		if err != nil {
			controller.Logger.Error("Failed to update user session. Switch to regular login process", err.Error(), reqMeta)
			goto regularLogin
		}

		if payload.ID != user.ID {
			controller.Logger.Error(
				"Failed to update user session. Switch to regular login process",
				"Already logged-in user tries to login as another user",
				reqMeta,
			)
			goto regularLogin
		}

		accessToken, refreshToken, err := updateSession(ctx, nil, user, payload)
		if err != nil {
			controller.Logger.Error("Failed to update user session. Switch to regular login process", err.Error(), reqMeta)
			goto regularLogin
		}

		controller.Logger.Info("Updating user session: OK", reqMeta)

		ctx.SetCookie(newAuthCookie(refreshToken))

		controller.Logger.Info("Authenticating user '"+body.Login+"': OK", reqMeta)

		return ctx.JSON(
			http.StatusOK,
			ResponseBody.Token{
				Message: "Пользователь успешно авторизован",
				AccessToken: accessToken.String(),
				ExpiresIn: int(accessToken.TTL()) / 1000,
			},
		)
	}
	regularLogin:

	deviceID, browser, err := getDeviceIDAndBrowser(ctx)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	// If session with this device is already exists and user tries to login from the same browser
	// (this is guard against cases when auth cookie was lost for some reasons)
	session, err := DB.Database.GetSessionByDeviceAndUserID(deviceID, user.ID)
	if err == nil && session.Browser == browser {
		// TODO code inside this block is very similar with the one that is several lines above, try to fix that
		controller.Logger.Info("Already existing user session was found for the specified device. Proceeding with it", reqMeta)
		controller.Logger.Info("Updating user session...", reqMeta)

		accessToken, refreshToken, err := updateSession(ctx, session, user, &UserDTO.Payload{
			ID: user.ID,
			Login: user.Login,
			Roles: user.Roles,
			SessionID: session.ID,
			Version: user.Version,
		})
		if err == nil {
			controller.Logger.Info("Updating user session: OK", reqMeta)

			ctx.SetCookie(newAuthCookie(refreshToken))

			controller.Logger.Info("Authenticating user '"+body.Login+"': OK", reqMeta)

			return ctx.JSON(
				http.StatusOK,
				ResponseBody.Token{
					Message: "Пользователь успешно авторизован",
					AccessToken: accessToken.String(),
					ExpiresIn: int(accessToken.TTL()) / 1000,
				},
			)
		}
	}

    payload := &UserDTO.Payload{
        ID: user.ID,
        Login: user.Login,
        Roles: user.Roles,
		SessionID: uuid.NewString(),
		Version: user.Version,
    }

    accessToken, refreshToken, err := token.NewAuthTokens(payload)
    if err != nil {
		controller.Logger.Error("Failed to authenticate user '"+body.Login+"'", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

	session, err = createSession(ctx, payload.SessionID, user.ID, config.Auth.RefreshTokenTTL())
	if err != nil {
		controller.Logger.Error("Failed to create session for user " + user.ID, err.Error(), reqMeta)
		return controller.ConvertErrorStatusToHTTP(err)
	}

	if err := DB.Database.SaveSession(session); err != nil {
		e := Error.NewStatusError(
			"Failed to save session",
			http.StatusInternalServerError,
		)
		controller.Logger.Error("Failed to login", err.Error(), reqMeta)
		return controller.ConvertErrorStatusToHTTP(e)
	}

    ctx.SetCookie(newAuthCookie(refreshToken))

	controller.Logger.Info("Authenticating user '"+body.Login+"': OK", reqMeta)

    return ctx.JSON(
        http.StatusOK,
        ResponseBody.Token{
            Message: "Пользователь успешно авторизован",
            AccessToken: accessToken.String(),
            ExpiresIn: int(accessToken.TTL()) / 1000,
        },
    )
}

func Logout(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

    authCookie, e := controller.GetAuthCookie(ctx)
    if e != nil {
        return e
    }

	refreshToken, err := controller.GetRefreshToken(ctx)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	claims := refreshToken.Claims.(jwt.MapClaims)

	payload, err := UserMapper.PayloadFromClaims(claims)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	sessionID := payload.SessionID

	if id := ctx.Param("sessionID"); id != "" {
		if e := validation.UUID(id); e != nil {
			return e.ToStatus(
				"User ID is missing", // this is not possible cuz uid already isn't empty stirng, but anyway...
				"User has invalid format (expected UUID)",
			)
		}
		sessionID = id
	}

	user, err := DB.Database.FindUserBySessionID(sessionID)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	act, err := UserMapper.TargetedActionDTOFromClaims(user.ID, claims)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	controller.Logger.Info("Logoutting user "+user.ID+"...", reqMeta)

	if err := DB.Database.RevokeSession(act, sessionID); err != nil {
		controller.Logger.Error("Failed to logout user "+user.ID, err.Error(), reqMeta)
		return controller.ConvertErrorStatusToHTTP(err)
	}

	if act.TargetUID == act.RequesterUID {
		controller.DeleteCookie(ctx, authCookie)
	}

	controller.Logger.Info("Logoutting user "+user.ID+": OK", reqMeta)

    return ctx.NoContent(http.StatusOK)
}

func Refresh(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	controller.Logger.Info("Refreshing auth tokens...", reqMeta)

    currentRefreshToken, err := controller.GetRefreshToken(ctx)
    if err != nil {
		controller.Logger.Error("Failed to refresh auth tokens", err.Error(), reqMeta)
        return controller.HandleTokenError(ctx, err)
    }

	payload, err := UserMapper.PayloadFromClaims(currentRefreshToken.Claims.(jwt.MapClaims))
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	user, err := DB.Database.FindUserByID(payload.ID)
	if err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	controller.Logger.Info("Updating user session...", reqMeta)

	accessToken, refreshToken, err := updateSession(ctx, nil, user, payload)
    if err != nil {
		controller.Logger.Error("Failed to update user session", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Info("Updating user session: OK", reqMeta)

    ctx.SetCookie(newAuthCookie(refreshToken))

	controller.Logger.Info("Refreshing auth tokens: OK", reqMeta)

    return ctx.JSON(
        http.StatusOK,
        ResponseBody.Token{
            Message: "Токены успешно обновлены",
            AccessToken: accessToken.String(),
            ExpiresIn: int(accessToken.TTL()) / 1000,
        },
    )
}

func Verify(ctx echo.Context) error {
	reqMeta := request.GetMetadata(ctx)

	controller.Logger.Info("Verifying access token...", reqMeta)

	// If token is invalid (expired, malformed etc) then this method will return error
    accessToken, err := controller.GetAccessToken(ctx)
    if err != nil {
		controller.Logger.Error("Verifying access token: ERROR", err.Error(), reqMeta)
        return controller.HandleTokenError(ctx, err)
    }

    payload, err := UserMapper.PayloadFromClaims(accessToken.Claims.(jwt.MapClaims))
    if err != nil {
		controller.Logger.Error("Failed to verify access token", err.Error(), reqMeta)
        return controller.ConvertErrorStatusToHTTP(err)
    }

	controller.Logger.Info("Verifying access token: OK", reqMeta)

    return ctx.JSON(http.StatusOK, payload)
}

