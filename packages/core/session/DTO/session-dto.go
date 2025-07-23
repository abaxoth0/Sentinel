package sessiondto

import "time"

// TODO Use composition instead of duplicating fields for each DTO? (the same goes for all DTOs in project)

type Full struct {
	ID 				  string	`json:"id" example:"254be108-2a12-4b0f-b095-c10cd80ef91d"`
	UserID 			  string	`json:"user-id" example:"7ee80427-b0c6-4120-b874-ba8567576b6d"`
	UserAgent 		  string	`json:"user-agent" example:"Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0"`
	// TODO use net.IP instead?
	IpAddress 		  string	`json:"ip-address" example:"8.8.8.8"`
	DeviceID 		  string	`json:"device-id" example:"Linux x86_64 Firefox/138.0"`
	DeviceType 		  string	`json:"device-type" example:"desktop"`
	OS 				  string	`json:"os" example:"Linux"`
	OSVersion		  string	`json:"os-version" example:"Unknown"`
	Browser 		  string	`json:"browser" example:"Firefox"`
	BrowserVersion	  string	`json:"browser-version" example:"138.0"`
	CreatedAt 		  time.Time	`json:"created-at" example:"2025-07-20T23:54:14.503Z"`
	LastUsedAt 		  time.Time	`json:"last-used-at" example:"2025-07-20T23:54:14.503Z"`
	ExpiresAt 		  time.Time	`json:"expires-at" example:"2025-07-20T23:54:14.503Z"`
	RevokedAt 		  time.Time	`json:"revoked-at" example:"2025-07-20T23:54:14.503Z"`
}

func (dto *Audit) IsRevoked() bool {
	return !dto.RevokedAt.IsZero()
}

type Public struct {
	ID 				  string	`json:"id" example:"254be108-2a12-4b0f-b095-c10cd80ef91d"`
	UserAgent 		  string	`json:"user-agent" example:"Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0"`
	IpAddress 		  string	`json:"ip-address" example:"8.8.8.8"`
	DeviceID 		  string	`json:"device-id" example:"Linux x86_64 Firefox/138.0"`
	DeviceType 		  string	`json:"device-type" example:"desktop"`
	OS 				  string	`json:"os" example:"Linux"`
	OSVersion		  string	`json:"os-version" example:"Unknown"`
	Browser 		  string	`json:"browser" example:"Firefox"`
	BrowserVersion	  string	`json:"browser-version" example:"138.0"`
	CreatedAt 		  time.Time	`json:"created-at" example:"2025-07-20T23:54:14.503Z"`
	LastUsedAt 		  time.Time	`json:"last-used-at" example:"2025-07-20T23:54:14.503Z"`
	ExpiresAt 		  time.Time	`json:"expires-at" example:"2025-07-20T23:54:14.503Z"`
}

type Audit struct {
	ChangedSessionID 	string 		`json:"changed-session-id"`
	ChangedByUserID 	string 		`json:"changed-by-user-id"`
	Operation			string		`json:"operation"`
	ChangedAt			time.Time	`json:"changed-at"`
	Reason				string		`json:"reason,omitempty"`

	*Full
}

