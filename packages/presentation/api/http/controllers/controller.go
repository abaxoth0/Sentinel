package controller

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"sentinel/packages/common/util"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/request"
	datamodel "sentinel/packages/presentation/data"

	"github.com/labstack/echo/v4"
)

var Logger = logger.NewSource("CONTROLLER", logger.Default)

func BindAndValidate[T datamodel.RequestValidator](ctx echo.Context, dest T) error {
    reqMeta, e := request.GetLogMeta(ctx)
	if e != nil {
		Logger.Panic("Failed to get log meta for the request", e.Error(), nil)
		return e
	}

    Logger.Trace("Binding and validating request...", reqMeta)

    if err := ctx.Bind(&dest); err != nil {
        Logger.Error("Failed to bind request", err.Error(), reqMeta)
        return err
    }

    if err := dest.Validate(); err != nil {
        Logger.Error("Request validation failed", err.Error(), reqMeta)
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    Logger.Trace("Binding and validating request: OK", reqMeta)

    return nil
}

type wwwAuthenticateParamas struct {
    Realm string
    Error string
    ErrorDescription string
}

func applyWWWAuthenticate(ctx echo.Context, params *wwwAuthenticateParamas) {
    ctx.Response().Header().Set(
        "WWW-Authenticate",
        `Bearer realm="`+params.Realm+`", error="`+params.Error+`", error_description="`+params.ErrorDescription+`"`,
    )
}

func HandleTokenError(ctx echo.Context, err *Error.Status) *echo.HTTPError {
	reqMeta, e := request.GetLogMeta(ctx)
	if e != nil {
		Logger.Panic("Failed to get log meta for the request", e.Error(), nil)
		return echo.NewHTTPError(http.StatusInternalServerError, e.Error())
	}

	Logger.Trace("Handling token error...", reqMeta)

    // token persist, but invalid
    if token.IsTokenError(err) {
        applyWWWAuthenticate(ctx, &wwwAuthenticateParamas{
            Realm: "api",
            Error: util.Ternary(err == token.TokenExpired, "expired_token", "invalid_token"),
            ErrorDescription: err.Error(),
        })

        authCookie, err := GetAuthCookie(ctx)
        if err != nil {
			Logger.Trace("Handling token error: OK", reqMeta)
            return err
        }

        DeleteCookie(ctx, authCookie)
        // token is missing
    } else if err == Error.StatusUnauthorized {
        applyWWWAuthenticate(ctx, &wwwAuthenticateParamas{
            Realm: "api",
            Error: "invalid_request",
            ErrorDescription: "No token provided",
        })
    }

	Logger.Trace("Handling token error: OK", reqMeta)

    return ConvertErrorStatusToHTTP(err)
}

