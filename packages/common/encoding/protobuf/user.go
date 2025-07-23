package pbencoding

import (
	pbgen "sentinel/packages/common/proto/generated"
	UserDTO "sentinel/packages/core/user/DTO"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func MarshallPublicUserDTO(dto *UserDTO.Public) ([]byte, error) {
	return marshall(&pbgen.PublicUserDTO{
		Id: dto.ID,
		Login: dto.Login,
		Roles: dto.Roles,
		DeletedAt: timestamppb.New(*dto.DeletedAt),
		Version: dto.Version,
	})
}

func UnmarshallPublicUserDTO(rawDTO []byte) (*UserDTO.Public, error) {
	dto, err := unmarshall(new(pbgen.PublicUserDTO), rawDTO)
	if err != nil {
		return nil, err
	}

	deletedAt := dto.DeletedAt.AsTime()

	return &UserDTO.Public{
		ID: dto.Id,
		Login: dto.Login,
		Roles: dto.Roles,
		DeletedAt: &deletedAt,
		Version: dto.Version,
	}, nil
}

func MarshallBasicUserDTO(dto *UserDTO.Basic) ([]byte, error) {
	return marshall(&pbgen.BasicUserDTO{
		Id: dto.ID,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: timestamppb.New(dto.DeletedAt),
		Version: dto.Version,
	})
}

func UnmarshallBasicUserDTO(rawDTO []byte) (*UserDTO.Basic, error) {
	dto, err := unmarshall(new(pbgen.BasicUserDTO), rawDTO)
	if err != nil {
		return nil, err
	}

	return &UserDTO.Basic{
		ID: dto.Id,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: dto.DeletedAt.AsTime(),
		Version: dto.Version,
	}, nil
}

func MarshallExtendedUserDTO(dto *UserDTO.Full) ([]byte, error) {
	return marshall(&pbgen.ExtendedUserDTO{
		Id: dto.ID,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: timestamppb.New(dto.DeletedAt),
		CreatedAt: timestamppb.New(dto.CreatedAt),
		Version: dto.Version,
	})
}

func UnmarshallExtendedUserDTO(rawDTO []byte) (*UserDTO.Full, error) {
	dto, err := unmarshall(new(pbgen.ExtendedUserDTO), rawDTO)
	if err != nil {
		return nil, err
	}

	return &UserDTO.Full{
		ID: dto.Id,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: dto.DeletedAt.AsTime(),
		CreatedAt: dto.CreatedAt.AsTime(),
		Version: dto.Version,
	}, nil
}

func MarshallAuditUserDTO(dto *UserDTO.Audit) ([]byte, error) {
	return marshall(&pbgen.AuditUserDTO{
		Id: dto.ID,
		ChangedById: dto.ChangedByUserID,
		ChangedUserId: dto.ChangedUserID,
		Operation: dto.Operation,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: timestamppb.New(dto.DeletedAt),
		ChangedAt: timestamppb.New(dto.ChangedAt),
		Version: dto.Version,
	})
}

func UnmarshallAuditUserDTO(rawDTO []byte) (*UserDTO.Audit, error) {
	dto, err := unmarshall(new(pbgen.AuditUserDTO), rawDTO)
	if err != nil {
		return nil, err
	}

	return &UserDTO.Audit{
		ChangedByUserID: dto.ChangedById,
		ChangedUserID: dto.ChangedUserId,
		Operation: dto.Operation,
		ChangedAt: dto.ChangedAt.AsTime(),
		Basic: &UserDTO.Basic{
			ID: dto.Id,
			Login: dto.Login,
			Password: dto.Password,
			Roles: dto.Roles,
			DeletedAt: dto.DeletedAt.AsTime(),
			Version: dto.Version,
		},
	}, nil
}

