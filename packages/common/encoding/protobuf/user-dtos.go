package pbencoding

import (
	"errors"
	"fmt"
	"sentinel/packages/common/encoding"
	pbgen "sentinel/packages/common/proto/generated"
	UserDTO "sentinel/packages/core/user/DTO"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func marshallUserDTO(pb any) ([]byte, error){
	encoding.Logger.Trace("Marshalling user DTO into a protobuf...", nil)

	var message proto.Message

	switch v := pb.(type) {
	case *pbgen.PublicUserDTO, *pbgen.BasicUserDTO, *pbgen.ExtendedUserDTO, *pbgen.AuditUserDTO:
		message = v.(proto.Message)
	default:
		return nil, fmt.Errorf(
			"[ TYPE ERROR ] Failed to marshall user DTO into protobuf: expected *pbgen.PublicUserDTO or *pbgen.BasicUserDTO or *pbgen.ExtendedUserDTO or *pbgen.AuditUserDTO, but got: %T", v)
	}

	data, err := proto.Marshal(message)
	if err != nil {
		return nil, errors.New("Faield to marshall user DTO in protobuf: " + err.Error())
	}

	encoding.Logger.Trace("Marshalling user DTO into a protobuf: OK", nil)

	return data, nil
}

func unmarshall[T proto.Message](message T, rawDTO []byte) (T, error) {
	encoding.Logger.Trace("Unmarshalling user DTO from protobuf...", nil)

	var zero T

	if err := proto.Unmarshal(rawDTO, message); err != nil {
		return zero, errors.New("Faield to decode public user DTO: " + err.Error())
	}

	encoding.Logger.Trace("Unmarshalling user DTO from protobuf: OK", nil)

	return message, nil
}

func MarshallPublicUserDTO(dto *UserDTO.Public) ([]byte, error) {
	return marshallUserDTO(&pbgen.PublicUserDTO{
		Id: dto.ID,
		Login: dto.Login,
		Roles: dto.Roles,
		DeletedAt: timestamppb.New(*dto.DeletedAt),
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
	}, nil
}

func MarshallBasicUserDTO(dto *UserDTO.Basic) ([]byte, error) {
	return marshallUserDTO(&pbgen.BasicUserDTO{
		Id: dto.ID,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: timestamppb.New(dto.DeletedAt),
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
	}, nil
}

func MarshallExtendedUserDTO(dto *UserDTO.Extended) ([]byte, error) {
	return marshallUserDTO(&pbgen.ExtendedUserDTO{
		Id: dto.ID,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: timestamppb.New(dto.DeletedAt),
		CreatedAt: timestamppb.New(dto.CreatedAt),
	})
}

func UnmarshallExtendedUserDTO(rawDTO []byte) (*UserDTO.Extended, error) {
	dto, err := unmarshall(new(pbgen.ExtendedUserDTO), rawDTO)
	if err != nil {
		return nil, err
	}
	return &UserDTO.Extended{
		ID: dto.Id,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: dto.DeletedAt.AsTime(),
		CreatedAt: dto.CreatedAt.AsTime(),
	}, nil
}

func MarshallAuditUserDTO(dto *UserDTO.Audit) ([]byte, error) {
	return marshallUserDTO(&pbgen.AuditUserDTO{
		Id: dto.ID,
		ChangedById: dto.ChangedByUserID,
		ChangedUserId: dto.ChangedUserID,
		Operation: dto.Operation,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: timestamppb.New(dto.DeletedAt),
		ChangedAt: timestamppb.New(dto.ChangedAt),
	})
}

func UnmarshallAuditUserDTO(rawDTO []byte) (*UserDTO.Audit, error) {
	dto, err := unmarshall(new(pbgen.AuditUserDTO), rawDTO)
	if err != nil {
		return nil, err
	}
	return &UserDTO.Audit{
		ID: dto.Id,
		ChangedByUserID: dto.ChangedById,
		ChangedUserID: dto.ChangedUserId,
		Operation: dto.Operation,
		Login: dto.Login,
		Password: dto.Password,
		Roles: dto.Roles,
		DeletedAt: dto.DeletedAt.AsTime(),
		ChangedAt: dto.ChangedAt.AsTime(),
	}, nil
}

