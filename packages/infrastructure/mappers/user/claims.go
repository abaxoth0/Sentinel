package usermapper

import (
	"fmt"
	"log"
	"sentinel/packages/common/util"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/token"

	Error "sentinel/packages/common/errors"

	"github.com/golang-jwt/jwt"
)

func convertToStrSlice(input []any) ([]string, error) {
    out := make([]string, len(input))

    for i,v := range input {
        str, ok := v.(string)
        if !ok {
            return nil, fmt.Errorf("element %d isn't a string", i)
        }

        out[i] = str
    }

    return out, nil
}

type claimsMapper[T any] func(claims jwt.MapClaims, roles []string) *T

func mapFromClaims[T UserDTO.Filter | UserDTO.Payload](
    claims jwt.MapClaims,
    mapper claimsMapper[T],
) (*T, *Error.Status) {
	if err := token.VerifyClaims(claims); err != nil {
		return nil, err
	}

    roles, err := convertToStrSlice(claims[token.UserRolesClaimsKey].([]any))
    if err != nil {
        log.Printf("[ ERROR ] Failed to create user filter from claims: %s\n", err.Error())
        return nil, Error.StatusInternalError
    }

    return mapper(claims, roles), nil
}

const NoTarget string = "no-targeted-user"

func FilterDTOFromClaims(targetUID string, claims jwt.MapClaims) (*UserDTO.Filter, *Error.Status) {
	return mapFromClaims(claims, func(claims jwt.MapClaims, roles []string) *UserDTO.Filter{
        return &UserDTO.Filter{
            TargetUID:      util.Ternary(targetUID == NoTarget, NoTarget, targetUID),
            RequesterUID:   claims[token.UserIdClaimsKey].(string),
            RequesterRoles: roles,
        }
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

