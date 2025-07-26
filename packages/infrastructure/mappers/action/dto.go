package actionmapper

import (
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/token"
)

func TargetedActionDTOFromClaims(targetUID string, claims *token.Claims) (*ActionDTO.UserTargeted) {
	return ActionDTO.NewUserTargeted(targetUID, claims.Subject, claims.Roles)
}

func BasicActionDTOFromClaims(claims *token.Claims) (*ActionDTO.Basic) {
	return ActionDTO.NewBasic(claims.Subject, claims.Roles)
}

