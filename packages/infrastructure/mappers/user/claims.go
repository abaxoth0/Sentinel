package usermapper

import (
	"errors"
	ActionDTO "sentinel/packages/core/action/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/mappers"
	"sentinel/packages/infrastructure/token"
	"strconv"

	Error "sentinel/packages/common/errors"

	"github.com/golang-jwt/jwt/v5"
)

func convertToStrSlice(input []any) ([]string, error) {
    out := make([]string, len(input))

    for i,v := range input {
        str, ok := v.(string)
        if !ok {
            idx := strconv.FormatInt(int64(i), 64)
            return nil, errors.New("Type error: element "+idx+" isn't a string")
        }

        out[i] = str
    }

    return out, nil
}

type claimsMapper[T any] func(claims jwt.MapClaims, roles []string) *T

func mapFromClaims[T ActionDTO.Targeted | ActionDTO.Basic | UserDTO.Payload](
    claims jwt.MapClaims,
    mapper claimsMapper[T],
) (*T, *Error.Status) {
	if err := token.VerifyClaims(claims); err != nil {
		return nil, err
	}

    roles, err := convertToStrSlice(claims[token.UserRolesClaimsKey].([]any))
    if err != nil {
        mappers.Logger.Error(
            "Failed to create user filter from claims",
            err.Error(),
        )
        return nil, Error.StatusInternalError
    }

    return mapper(claims, roles), nil
}

func TargetedActionDTOFromClaims(targetUID string, claims jwt.MapClaims) (*ActionDTO.Targeted, *Error.Status) {
	return mapFromClaims(claims, func(claims jwt.MapClaims, roles []string) *ActionDTO.Targeted{
        return ActionDTO.NewTargeted(
            targetUID,
            claims[token.UserIdClaimsKey].(string),
            roles,
        )
    })
}


func BasicActionDTOFromClaims(claims jwt.MapClaims) (*ActionDTO.Basic, *Error.Status) {
	return mapFromClaims(claims, func(claims jwt.MapClaims, roles []string) *ActionDTO.Basic{
        return ActionDTO.NewBasic(
            claims[token.UserIdClaimsKey].(string),
            roles,
        )
    })
}

// TODO receive jwt.Claims instead of MapClaims (for all of that funcs)

// IMPORTANT: Use this function only if token is valid.
func PayloadFromClaims(claims jwt.MapClaims) (*UserDTO.Payload, *Error.Status) {
	return mapFromClaims(claims, func(claims jwt.MapClaims, roles []string) *UserDTO.Payload{
        return &UserDTO.Payload{
            ID:    claims[token.UserIdClaimsKey].(string),
            Login: claims[token.UserLoginClaimsKey].(string),
            Roles: roles,
        }
    })
}

