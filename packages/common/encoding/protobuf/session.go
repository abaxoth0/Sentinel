package pbencoding

import (
	pbgen "sentinel/packages/common/proto/generated"
	SessionDTO "sentinel/packages/core/session/DTO"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func MarshallFullSessionDTO(dto *SessionDTO.Full) ([]byte, error) {
	return marshall(&pbgen.FullSessionDTO{
		ID: dto.ID,
		UserID: dto.UserID,
		UserAgent: dto.UserAgent,
		IpAddress: dto.IpAddress,
		DeviceID: dto.DeviceID,
		DeviceType: dto.DeviceType,
		OS: dto.OS,
		OSVersion: dto.OSVersion,
		Browser: dto.Browser,
		BrowserVersion: dto.BrowserVersion,
		CreatedAt: timestamppb.New(dto.CreatedAt),
		LastUsedAt: timestamppb.New(dto.LastUsedAt),
		ExpiresAt: timestamppb.New(dto.ExpiresAt),
		RevokedAt: timestamppb.New(dto.RevokedAt),
	})
}

func UnmarshallFullSessionDTO(rawDTO []byte) (*SessionDTO.Full, error) {
	dto, err := unmarshall(new(pbgen.FullSessionDTO), rawDTO)
	if err != nil {
		return nil, err
	}

	return &SessionDTO.Full{
		ID: dto.ID,
		UserID: dto.UserID,
		UserAgent: dto.UserAgent,
		IpAddress: dto.IpAddress,
		DeviceID: dto.DeviceID,
		DeviceType: dto.DeviceType,
		OS: dto.OS,
		OSVersion: dto.OSVersion,
		Browser: dto.Browser,
		BrowserVersion: dto.BrowserVersion,
		CreatedAt: dto.CreatedAt.AsTime(),
		LastUsedAt: dto.LastUsedAt.AsTime(),
		ExpiresAt: dto.ExpiresAt.AsTime(),
		RevokedAt: dto.RevokedAt.AsTime(),
	}, nil
}

func MarshallPublicSessionDTO(dto *SessionDTO.Public) ([]byte, error) {
	return marshall(&pbgen.PublicSessionDTO{
		ID: dto.ID,
		UserAgent: dto.UserAgent,
		IpAddress: dto.IpAddress,
		DeviceID: dto.DeviceID,
		DeviceType: dto.DeviceType,
		OS: dto.OS,
		OSVersion: dto.OSVersion,
		Browser: dto.Browser,
		BrowserVersion: dto.BrowserVersion,
		CreatedAt: timestamppb.New(dto.CreatedAt),
		LastUsedAt: timestamppb.New(dto.LastUsedAt),
		ExpiresAt: timestamppb.New(dto.ExpiresAt),
	})
}

func UnmarshallPublicSessionDTO(rawDTO []byte) (*SessionDTO.Public, error) {
	dto, err := unmarshall(new(pbgen.PublicSessionDTO), rawDTO)
	if err != nil {
		return nil, err
	}

	return &SessionDTO.Public{
		ID: dto.ID,
		UserAgent: dto.UserAgent,
		IpAddress: dto.IpAddress,
		DeviceID: dto.DeviceID,
		DeviceType: dto.DeviceType,
		OS: dto.OS,
		OSVersion: dto.OSVersion,
		Browser: dto.Browser,
		BrowserVersion: dto.BrowserVersion,
		CreatedAt: dto.CreatedAt.AsTime(),
		LastUsedAt: dto.LastUsedAt.AsTime(),
		ExpiresAt: dto.ExpiresAt.AsTime(),
	}, nil
}

