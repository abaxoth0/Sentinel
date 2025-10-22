package controller

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"sentinel/packages/common/util"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/cookie"
	"sentinel/packages/presentation/api/http/request"
	RequestBody "sentinel/packages/presentation/data/request"

	"github.com/labstack/echo/v4"
)

var Log = logger.NewSource("CONTROLLER", logger.Default)

func BindAndValidate[T RequestBody.Validator](ctx echo.Context, dest T) error {
	reqMeta := request.GetMetadata(ctx)

	Log.Trace("Binding and validating request...", reqMeta)

	if err := ctx.Bind(&dest); err != nil {
		Log.Error("Failed to bind request", err.Error(), reqMeta)
		return err
	}

	if err := dest.Validate(); err != nil {
		Log.Error("Request validation failed", err.Error(), reqMeta)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	Log.Trace("Binding and validating request: OK", reqMeta)

	return nil
}

type wwwAuthenticateParamas struct {
	Realm            string
	Error            string
	ErrorDescription string
}

func applyWWWAuthenticate(ctx echo.Context, params *wwwAuthenticateParamas) {
	ctx.Response().Header().Set(
		"WWW-Authenticate",
		`Bearer realm="`+params.Realm+`", error="`+params.Error+`", error_description="`+params.ErrorDescription+`"`,
	)
}

func HandleTokenError(ctx echo.Context, err *Error.Status) *Error.Status {
	reqMeta := request.GetMetadata(ctx)

	Log.Trace("Handling token error...", reqMeta)

	// token persist, but invalid
	if token.IsTokenError(err) {
		applyWWWAuthenticate(ctx, &wwwAuthenticateParamas{
			Realm:            "api",
			Error:            util.Ternary(err == token.TokenExpired, "expired_token", "invalid_token"),
			ErrorDescription: err.Error(),
		})

		authCookie, err := cookie.GetAuthCookie(ctx)
		if err != nil {
			Log.Trace("Handling token error: OK", reqMeta)
			return err
		}

		cookie.DeleteCookie(ctx, authCookie)
		// token is missing
	} else if err == Error.StatusUnauthorized {
		applyWWWAuthenticate(ctx, &wwwAuthenticateParamas{
			Realm:            "api",
			Error:            "invalid_request",
			ErrorDescription: "No token provided",
		})
	}

	Log.Trace("Handling token error: OK", reqMeta)

	return err
}
