package oauthcontroller

import (
	"crypto/ed25519"
	"net/http"
	"sentinel/packages/common/config"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	RequestBody "sentinel/packages/presentation/data/request"
	ResponseBody "sentinel/packages/presentation/data/response"
	"github.com/labstack/echo/v4"
)

var isInit = false

func Init() {
	if isInit {
		controller.Log.Panic("Failed to initialize OAuth controller", "Controller already initialized", nil)
		return
	}

	initGoogle()

	isInit = true
}

// @Summary 		OAuth 2.0 Token Introspection
// @Description 	RFC 7662 (https://datatracker.ietf.org/doc/html/rfc7662). Valid token types are: access, refresh and activate.
// @ID 				oauth-introspect
// @Tags			oauth
// @Param 			Token body requestbody.Introspect true "OAuth2.0 token which must be introspected"
// @Accept			json
// @Produce			json
// @Success			200 			{object} 	responsebody.Introspection
// @Failure			400,401,500 	{object} 	responsebody.Error
// @Router			/auth/oauth/introspect [post]
// @Security		BearerAuth
func IntrospectOAuthToken(ctx echo.Context) error {
	act := controller.GetBasicAction(ctx)

	if err := authz.User.OAuthIntrospect(act.RequesterRoles); err != nil {
		return controller.ConvertErrorStatusToHTTP(err)
	}

	body := RequestBody.Introspect{}

	if err := controller.BindAndValidate(ctx, &body); err != nil {
		return err
	}

	var key ed25519.PublicKey

	switch body.Type {
	case "access":
		key = config.Secret.AccessTokenPublicKey
	case "refresh":
		key = config.Secret.RefreshTokenPublicKey
	case "activate":
		key = config.Secret.ActivationTokenPublicKey
	default:
		return echo.NewHTTPError(
			http.StatusBadRequest,
			`Invalid token type, valid types are: "access", "refresh" and "activate". But got: ` + body.Type,
		)
	}

	tk, err := token.ParseSingedToken(body.Token, key)
	if err != nil {
		return echo.NewHTTPError(err.Status(), "Failed to parse specified token: " + err.Error())
	}

	claims := tk.Claims.(*token.Claims)

	return ctx.JSON(http.StatusOK, ResponseBody.Introspection{
		Active: 	true,
		SessionID: 	claims.ID,
		Subject: 	claims.Subject,
		Issuer: 	claims.Issuer,
		Audience: 	claims.Audience,
		ExpiresAt: 	claims.ExpiresAt.Unix(),
		IssuedAt: 	claims.IssuedAt.Unix(),
		Scope: 		claims.Roles,
	})
}

