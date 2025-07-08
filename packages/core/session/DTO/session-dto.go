package sessiondto

import "time"

type Full struct {
	ID 				  string
	UserID 			  string
	UserAgent 		  string
	IpAddress 		  string
	DeviceID 		  string
	DeviceType 		  string
	OS 				  string
	OSVersion		  string
	Browser 		  string
	BrowserVersion	  string
	Location 		  string
	CreatedAt 		  time.Time
	LastUsedAt 		  time.Time
	ExpiresAt 		  time.Time
	Revoked 		  bool
}

type DevicePayload struct {
	Fingerprint string
	Name 		string
	Type 		string
}

type NetworkPayload struct {
	UserAgent string
	IpAddres  string
	location  string
}

