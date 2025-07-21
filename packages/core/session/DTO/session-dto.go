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
	CreatedAt 		  time.Time
	LastUsedAt 		  time.Time
	ExpiresAt 		  time.Time
	Revoked 		  bool
}

type Public struct {
	ID 				  string	`json:"id" example:"254be108-2a12-4b0f-b095-c10cd80ef91d"`
	UserAgent 		  string	`json:"user-agent" example:"Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0"`
	IpAddress 		  string	`json:"ip-address" example:"8.8.8.8"`
	DeviceID 		  string	`json:"device-id" example:"Linux x86_64 Firefox/138.0:`
	DeviceType 		  string	`json:"device-type" example:"desktop"`
	OS 				  string	`json:"os" example:"Linux"`
	OSVersion		  string	`json:"os-version" example:"Unknown"`
	Browser 		  string	`json:"browser" example:"Firefox"`
	BrowserVersion	  string	`json:"browser-version" example:"138.0"`
	CreatedAt 		  time.Time	`json:"created-at" example:"2025-07-20T23:54:14.503Z"`
	LastUsedAt 		  time.Time	`json:"last-used-at" example:"2025-07-20T23:54:14.503Z"`
	ExpiresAt 		  time.Time	`json:"expires-at" example:"2025-07-20T23:54:14.503Z"`
}

