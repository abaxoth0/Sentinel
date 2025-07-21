package locationdto

import (
	"net"
	"sentinel/packages/common/util"
	"time"
)

type Full struct {
	ID 			string		`json:"id" example:"0de6c6e9-5360-4cd8-a068-24ea035a0bd7"`
	SessionID	string		`json:"session-id" example:"c27ee824-a78c-47c7-ae53-bf15f73734b3"`
	IP 			net.IP		`json:"ip" example:"8.8.8.8"`
	// ISO 3166-1 alpha-2
	Country 	string		`json:"countryCode" example:"US"`
	// ISO 3166-2 region code
	Region		string		`json:"region" example:"VA"`
	City		string		`json:"city" example:"Ashburn"`
	Latitude 	float32		`json:"lat" example:"39.03"`
	Longitude 	float32		`json:"lon" example:"-77.5"`
	ISP 		string		`json:"isp" example:"Google LLC"`
	DeletedAt	time.Time	`json:"deleted-at" example:"2025-07-15T22:27:50.294Z"`
	CreatedAt	time.Time	`json:"created-at" example:"2025-07-15T22:27:50.294Z"`
}

// Creates new Public location DTO based on current Full DTO
func (dto *Full) MakePublic() *Public {
	deletedAt := util.Ternary(dto.DeletedAt.IsZero(), nil, &dto.DeletedAt)
	createdAt := util.Ternary(dto.CreatedAt.IsZero(), nil, &dto.CreatedAt)

	return &Public{
		ID: dto.ID,
		IP: dto.IP.To4().String(),
		Country: dto.Country,
		Region: dto.Region,
		City: dto.City,
		Latitude: dto.Latitude,
		Longitude: dto.Longitude,
		ISP: dto.ISP,
		DeletedAt: deletedAt,
		CreatedAt: createdAt,
	}
}

type Public struct {
	ID 			string		`json:"id" example:"0de6c6e9-5360-4cd8-a068-24ea035a0bd7"`
	IP 			string		`json:"ip" example:"8.8.8.8"`
	// ISO 3166-1 alpha-2
	Country 	string		`json:"countryCode" example:"US"`
	// ISO 3166-2 region code
	Region		string		`json:"region" example:"VA"`
	City		string		`json:"city" example:"Ashburn"`
	Latitude 	float32		`json:"lat" example:"39.03"`
	Longitude 	float32		`json:"lon" example:"-77.5"`
	ISP 		string		`json:"isp" example:"Google LLC"`
	DeletedAt	*time.Time	`json:"deleted-at,omitempty" example:"2025-07-15T22:27:50.294Z"`
	CreatedAt	*time.Time	`json:"created-at,omitempty" example:"2025-07-15T22:27:50.294Z"`
}

