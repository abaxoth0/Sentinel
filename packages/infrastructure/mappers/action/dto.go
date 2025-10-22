package actionmapper

import (
	ActionDTO "sentinel/packages/core/action/DTO"
	"sentinel/packages/infrastructure/token"
)

func TargetedActionDTOFromClaims(targetUID string, claims *token.Claims) *ActionDTO.UserTargeted {
	return ActionDTO.NewUserTargeted(targetUID, claims.Subject, claims.Roles)
}
