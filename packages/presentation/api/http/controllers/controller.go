package controller

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"sentinel/packages/common/util"
	"sentinel/packages/infrastructure/token"
	datamodel "sentinel/packages/presentation/data"

	"github.com/labstack/echo/v4"
)

var Logger = logger.NewSource("CONTROLLER", logger.Default)

func RequestInfo(ctx echo.Context) string {
    req := ctx.Request()

    return " ("+req.RemoteAddr+" "+req.Method+" "+req.URL.Path+"; user agent: "+req.UserAgent()+")"
}

func BindAndValidate[T datamodel.RequestValidator](ctx echo.Context, dest T) error {
    reqInfo := RequestInfo(ctx)

    Logger.Trace("Binding and validating request..." + reqInfo)

    if err := ctx.Bind(&dest); err != nil {
        Logger.Error("Failed to bind request" + reqInfo, err.Error())
        return err
    }

    if err := dest.Validate(); err != nil {
        Logger.Error("Request validation failed" + reqInfo, err.Error())
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    Logger.Trace("Binding and validating request: OK" + reqInfo)

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
	reqInfo := RequestInfo(ctx)

	Logger.Trace("Handling token error..." + reqInfo)

    // token persist, but invalid
    if token.IsTokenError(err) {
        applyWWWAuthenticate(ctx, &wwwAuthenticateParamas{
            Realm: "api",
            Error: util.Ternary(err == token.TokenExpired, "expired_token", "invalid_token"),
            ErrorDescription: err.Error(),
        })

        authCookie, err := GetAuthCookie(ctx)
        if err != nil {
			Logger.Trace("Handling token error: OK" + reqInfo)
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

	Logger.Trace("Handling token error: OK" + reqInfo)

    return ConvertErrorStatusToHTTP(err)
}

