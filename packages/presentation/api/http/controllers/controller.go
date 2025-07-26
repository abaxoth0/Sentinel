package controller

import (
	"net/http"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"sentinel/packages/common/util"
	ActionDTO "sentinel/packages/core/action/DTO"
	ActionMapper "sentinel/packages/infrastructure/mappers/action"
	"sentinel/packages/infrastructure/token"
	"sentinel/packages/presentation/api/http/request"
	RequestBody "sentinel/packages/presentation/data/request"

	"github.com/labstack/echo/v4"
)

var Logger = logger.NewSource("CONTROLLER", logger.Default)

func BindAndValidate[T RequestBody.Validator](ctx echo.Context, dest T) error {
    reqMeta := request.GetMetadata(ctx)

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
	reqMeta := request.GetMetadata(ctx)

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

func newActionDTO[T ActionDTO.Any](
	ctx echo.Context,
	uid string,
	mapFunc func (uid string, claims *token.Claims) (T),
) (T, *echo.HTTPError) {
	var zero T
    reqMeta := request.GetMetadata(ctx)

    Logger.Trace("Retrieving access token from the request...", reqMeta)

    accessToken, err := GetAccessToken(ctx)
    if err != nil {
        Logger.Error("Failed to retrieve valid access token from the request", err.Error(), reqMeta)
        return zero, HandleTokenError(ctx, err)
    }

    Logger.Trace("Retrieving access token from the request: OK", reqMeta)
    Logger.Trace("Creating action DTO from token claims...", reqMeta)

	// claims can be trusted if token is valid
	act := mapFunc(uid, accessToken.Claims.(*token.Claims))

    Logger.Trace("Creating action DTO from token claims: OK", reqMeta)

    return act, nil
}

func NewBasicActionDTO(ctx echo.Context) (*ActionDTO.Basic, *echo.HTTPError) {
	return newActionDTO(ctx, "", func (_ string, claims *token.Claims) (*ActionDTO.Basic) {
		return ActionMapper.BasicActionDTOFromClaims(claims)
	})
}

func NewTargetedActionDTO(ctx echo.Context, uid string) (*ActionDTO.UserTargeted, *echo.HTTPError) {
	return newActionDTO(ctx, uid, func (id string, claims *token.Claims) (*ActionDTO.UserTargeted) {
		return ActionMapper.TargetedActionDTOFromClaims(id, claims)
	})
}

