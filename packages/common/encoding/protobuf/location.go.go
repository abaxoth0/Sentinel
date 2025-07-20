package pbencoding

import (
	pbgen "sentinel/packages/common/proto/generated"
	LocationDTO "sentinel/packages/core/location/DTO"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func MarshallFullLocationDTO(dto *LocationDTO.Full) ([]byte, error) {
	return marshall(&pbgen.FullLocationDTO{
		ID: dto.ID,
		SessionID: dto.SessionID,
		IP: dto.IP,
		Country: dto.Country,
		Region: dto.Region,
		City: dto.City,
		Latitude: dto.Latitude,
		Longitude: dto.Longitude,
		ISP: dto.ISP,
		DeletedAt: timestamppb.New(dto.DeletedAt),
		CreatedAt: timestamppb.New(dto.CreatedAt),
	})
}

func UnmarshallFullLocationDTO(rawDTO []byte) (*LocationDTO.Full, error) {
	dto, err := unmarshall(new(pbgen.FullLocationDTO), rawDTO)
	if err != nil {
		return nil, err
	}

	return &LocationDTO.Full{
		ID: dto.ID,
		SessionID: dto.SessionID,
		IP: dto.IP,
		Country: dto.Country,
		Region: dto.Region,
		City: dto.City,
		Latitude: dto.Latitude,
		Longitude: dto.Longitude,
		ISP: dto.ISP,
		DeletedAt: dto.DeletedAt.AsTime(),
		CreatedAt: dto.CreatedAt.AsTime(),
	}, nil
}

