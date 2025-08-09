package sharedcontroller

import (
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authz"
	UserMapper "sentinel/packages/infrastructure/mappers/user"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/cookie"
	"sentinel/packages/presentation/api/http/request"
	ResponseBody "sentinel/packages/presentation/data/response"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func AuthenticateWithNewSession(ctx echo.Context, user *UserDTO.Full, audience []string) error {
	payload := &UserDTO.Payload{
		ID: user.ID,
		Login: user.Login,
		Roles: user.Roles,
		SessionID: uuid.NewString(),
		Version: user.Version,
		Audience: audience,
	}

	accessToken, refreshToken, err := token.NewAuthTokens(payload)
	if err != nil {
		return err
	}

	session, err := createSession(ctx, payload.SessionID, user.ID, config.Auth.RefreshTokenTTL())
	if err != nil {
		return err
	}

	if err := DB.Database.SaveSession(session); err != nil {
		return Error.NewStatusError("Failed to save session", http.StatusInternalServerError)
	}

	act := ActionDTO.NewUserTargeted(user.ID, user.ID, user.Roles)

	if err := updateOrCreateLocation(act, session.ID, session.IpAddress); err != nil {
		if e := DB.Database.RevokeSession(act, session.ID); e != nil {
			return e
		}
		return err
	}

	ctx.SetCookie(cookie.NewAuthCookie(refreshToken))

	controller.Log.Info("Authenticating user '"+user.Login+"': OK", request.GetMetadata(ctx))

	return ctx.JSON(
		http.StatusOK,
		ResponseBody.Token{
			Message: "Пользователь успешно авторизован",
			AccessToken: accessToken.String(),
			ExpiresIn: int(accessToken.TTL()) / 1000,
		},
	)
}

func Authenticate(ctx echo.Context, user *UserDTO.Full, audience []string) error {
	if tk, err := GetRefreshToken(ctx); err == nil {
		reqMeta := request.GetMetadata(ctx)

		payload := UserMapper.PayloadFromClaims(tk.Claims.(*token.Claims))

		if payload.ID != user.ID {
			controller.Log.Error(
				"Failed to update user session. Switch to regular login process",
				"Already logged-in user tries to login as another user",
				reqMeta,
			)
			goto regular_login
		}

		accessToken, refreshToken, err := UpdateSession(ctx, nil, user, payload)
		if err != nil {
			if err == authz.InsufficientPermissions || err == authz.DeniedByActionGatePolicy {
				return err
			}
			controller.Log.Error("Failed to update user session. Switch to regular login process", err.Error(), reqMeta)
			goto regular_login
		}

		ctx.SetCookie(cookie.NewAuthCookie(refreshToken))

		controller.Log.Info("Authenticating user '"+user.Login+"': OK", reqMeta)

		return ctx.JSON(
			http.StatusOK,
			ResponseBody.Token{
				Message: "Пользователь успешно авторизован",
				AccessToken: accessToken.String(),
				ExpiresIn: int(accessToken.TTL()) / 1000,
			},
		)
	}
	regular_login:

	deviceID, browser, err := getDeviceIDAndBrowser(ctx)
	if err != nil {
		return err
	}

	// If session with this device is already exists and user tries to login from the same browser
	// (this is guard against cases when auth cookie was lost for some reasons)
	session, e := DB.Database.GetSessionByDeviceAndUserID(deviceID, user.ID)
	if e == nil && session.Browser == browser {
		reqMeta := request.GetMetadata(ctx)

		controller.Log.Info("Already existing user session was found for the specified device. Proceeding with it", reqMeta)

		accessToken, refreshToken, err := UpdateSession(ctx, session, user, &UserDTO.Payload{
			ID: user.ID,
			Login: user.Login,
			Roles: user.Roles,
			SessionID: session.ID,
			Version: user.Version,
		})
		if err == nil {
			ctx.SetCookie(cookie.NewAuthCookie(refreshToken))

			controller.Log.Info("Authenticating user '"+user.Login+"': OK", reqMeta)

			ctx.SetCookie(cookie.NewAuthCookie(refreshToken))

			return ctx.JSON(
				http.StatusOK,
				ResponseBody.Token{
					Message: "Пользователь успешно авторизован",
					AccessToken: accessToken.String(),
					ExpiresIn: int(accessToken.TTL()) / 1000,
				},
			)
		}
		if err == authz.InsufficientPermissions || err == authz.DeniedByActionGatePolicy {
			return err
		}
	}

	return AuthenticateWithNewSession(ctx, user, audience)
}

