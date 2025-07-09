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

type Public struct {
	ID 				  string	`json:"id"`
	UserAgent 		  string	`json:"user-agent"`
	IpAddress 		  string	`json:"ip-address"`
	DeviceID 		  string	`json:"device-id"`
	DeviceType 		  string	`json:"device-type"`
	OS 				  string	`json:"os"`
	OSVersion		  string	`json:"os-version"`
	Browser 		  string	`json:"browser"`
	BrowserVersion	  string	`json:"browser-version"`
	Location 		  string	`json:"location"`
	CreatedAt 		  time.Time	`json:"created-at"`
	LastUsedAt 		  time.Time	`json:"last-used-at"`
	ExpiresAt 		  time.Time	`json:"expires-at"`
}

