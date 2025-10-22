package pbencoding

import (
	pbgen "sentinel/packages/common/proto/generated"
	"sentinel/packages/common/util"
	UserDTO "sentinel/packages/core/user/DTO"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func MarshallFullUserDTO(dto *UserDTO.Full) ([]byte, error) {
	return marshall(&pbgen.FullUserDTO{
		Id:        dto.ID,
		Login:     dto.Login,
		Password:  dto.Password,
		Roles:     dto.Roles,
		DeletedAt: timestamppb.New(util.SafeDereference(dto.DeletedAt)),
		CreatedAt: timestamppb.New(dto.CreatedAt),
		Version:   dto.Version,
	})
}

func UnmarshallFullUserDTO(rawDTO []byte) (*UserDTO.Full, error) {
	dto, err := unmarshall(new(pbgen.FullUserDTO), rawDTO)
	if err != nil {
		return nil, err
	}

	deletedAt := dto.DeletedAt.AsTime()

	return &UserDTO.Full{
		Basic: UserDTO.Basic{
			ID:        dto.Id,
			Login:     dto.Login,
			Password:  dto.Password,
			Roles:     dto.Roles,
			DeletedAt: util.Ternary(deletedAt.IsZero(), nil, &deletedAt),
			Version:   dto.Version,
		},
		CreatedAt: dto.CreatedAt.AsTime(),
	}, nil
}
