package controller

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	"sentinel/packages/infrastructure/token"
	datamodel "sentinel/packages/presentation/data"

	"github.com/labstack/echo/v4"
)

func BindAndValidate[T datamodel.RequestValidator](ctx echo.Context, dest T) error {
    if err := ctx.Bind(&dest); err != nil {
        return err
    }

    if err := dest.Validate(); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

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
    // token persist, but invalid
    if token.IsTokenError(err) {
        applyWWWAuthenticate(ctx, &wwwAuthenticateParamas{
            Realm: "api",
            Error: util.Ternary(err == token.TokenExpired, "expired_token", "invalid_token"),
            ErrorDescription: err.Error(),
        })

        authCookie, err := GetAuthCookie(ctx)
        if err != nil {
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
    return ConvertErrorStatusToHTTP(err)
}

