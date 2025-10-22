package cache

import Error "sentinel/packages/common/errors"

func BulkInvalidateBasicUserDTO(UIDs, logins []string) *Error.Status {
	keys := make([]string, 0, len(UIDs)*5+len(logins)*2)

	for _, id := range UIDs {
		keys = append(keys,
			KeyBase[AnyUserById]+id,
			KeyBase[UserById]+id,
			KeyBase[DeletedUserById]+id,
			KeyBase[UserRolesById]+id,
			KeyBase[UserVersionByID]+id,
		)
	}

	for _, login := range logins {
		keys = append(keys,
			KeyBase[AnyUserByLogin]+login,
			KeyBase[UserByLogin]+login,
		)
	}

	const batchSize = 500
	for i := 0; i < len(keys); i += batchSize {
		end := min(len(keys), i+batchSize)

		if err := Client.ProgressiveDelete(keys[i:end]); err != nil {
			return err
		}
	}

	return nil
}
