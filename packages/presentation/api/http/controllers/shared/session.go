package sharedcontroller

// 0xd34df00d
// 0xd4b4c507
// 0x0d34d10c
// 0x001e5a5a

import (
	"net"
	"net/http"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/validation"
	ActionDTO "sentinel/packages/core/action/DTO"
	SessionDTO "sentinel/packages/core/session/DTO"
	UserDTO "sentinel/packages/core/user/DTO"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/token"
	controller "sentinel/packages/presentation/api/http/controllers"
	"sentinel/packages/presentation/api/http/request"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mileusna/useragent"
)

type deviceType string

const (
	mobile 	deviceType = "mobile"
	tablet 	deviceType = "tablet"
	desktop deviceType = "desktop"
)

func getDeviceIDAndBrowser(ctx echo.Context) (deviceID string, browser string, err *Error.Status) {
	ua, err := getUserAgent(ctx)
	if err != nil {
		return "", "", err
	}

	_, id := getDeviceInfo(ua)

	// ua.Name is always will be non-empty string. This is guaranteed by getUserAgent()
	return id, ua.Name, nil
}

func getDeviceInfo(userAgent useragent.UserAgent) (deviceType deviceType, deviceID string) {
	if userAgent.Mobile {
		return mobile, getMobileDeviceID(userAgent)
	}
	if userAgent.Tablet {
		return tablet, getTableDeviceID(userAgent)
	}
	return desktop, getDesktopDeviceID(userAgent)
}

func getMobileDeviceID(ua useragent.UserAgent) string {
	id := ua.Device
	if id == "" {
		return "Unknown mobile"
	}
	return id
}

func getTableDeviceID(ua useragent.UserAgent) string {
	id := ua.Device
	if id == "" {
		return "Unknown tablet"
	}
	return id
}

func getDesktopDeviceID(ua useragent.UserAgent) string {
	os := ua.OS
	if os == "" {
		os = "Unknown OS "
	}
	return os + ua.Name
}

func getOS(ua useragent.UserAgent) (os, osVersion string) {
	os = ua.OS
	osVersion = ua.OSVersion

	if os == "" {
		os = "Unknown OS"
	}
	if osVersion == "" {
		osVersion = "Unknown Version"
	}

	return os, osVersion
}

func getUserAgent(ctx echo.Context) (useragent.UserAgent, *Error.Status) {
	userAgent := ctx.Request().UserAgent()

	if strings.ReplaceAll(userAgent, " ", "") == "" {
		return useragent.UserAgent{}, Error.NewStatusError(
			"User Agent is missing",
			http.StatusBadRequest,
		)
	}

	ua := useragent.Parse(userAgent)
	if ua.Name == "" {
		return useragent.UserAgent{}, Error.NewStatusError(
			"Invalid User Agent: browser isn't specified",
			http.StatusBadRequest,
		)
	}

	return ua, nil
}

func createSession(ctx echo.Context, ID string, UID string, ttl time.Duration) (*SessionDTO.Full, *Error.Status) {
	if err := validation.UUID(ID); err != nil {
		controller.Log.Panic(
			"Failed to create user session",
			err.ToStatus(
				"Session ID is missing",
				"Session ID has invalid format (UUID expected)",
			).Error(),
			request.GetMetadata(ctx),
		)
		return nil, Error.StatusInternalError
	}

	ua, err := getUserAgent(ctx)
	if err != nil {
		return nil, err
	}

	deviceType, deviceID := getDeviceInfo(ua)
	os, osVersion := getOS(ua)

	now := time.Now()

	return &SessionDTO.Full{
		ID: ID,
		UserID: UID,
		UserAgent: ua.String,
		IpAddress: net.ParseIP(ctx.RealIP()),
		DeviceID: deviceID,
		DeviceType: string(deviceType),
		OS: os,
		OSVersion: osVersion,
		Browser: ua.Name,
		BrowserVersion: ua.Version,
		CreatedAt: now,
		LastUsedAt: now,
		ExpiresAt: now.Add(ttl),
	}, nil
}

var deviceIDMismatch = Error.NewStatusError(
	"Detected new device ID",
	http.StatusConflict,
)
var browserMismatch = Error.NewStatusError(
	"Detected new browser",
	http.StatusConflict,
)
var osMismatch = Error.NewStatusError(
	"Detected new OS",
	http.StatusConflict,
)

func actualizeSession(
	ctx echo.Context,
	session *SessionDTO.Full,
	ttl time.Duration,
) (*SessionDTO.Full, *Error.Status) {
	ua, err := getUserAgent(ctx)
	if err != nil {
		return nil, err
	}

	_, deviceID := getDeviceInfo(ua)
	os, osVersion := getOS(ua)

	// Browser mismatch
	// Sessions mustn't be shared across them, so each there are must be new session for each browser
	if session.Browser != ua.Name {
		return nil, browserMismatch
	}
	// OS mismatch
	// In case with PC, each OS must be treated as new device
	if session.OS != os {
		return nil, osMismatch
	}
	// Device ID mismatch (for mobiles and tablets it's actual device model, for OS it's OS + browser name)
	// e.g. was some Android-based phone (samsung, for example), but new one is on iPhone or on some PC on Windows
	if session.DeviceID != deviceID {
		return nil, deviceIDMismatch
	}

	now := time.Now()

	return &SessionDTO.Full{
		ID: session.ID,
		UserID: session.UserID,
		UserAgent: ua.String,
		IpAddress: net.ParseIP(ctx.RealIP()),
		DeviceID: session.DeviceID,
		DeviceType: session.DeviceType,
		OS: session.OS,
		OSVersion: osVersion,
		Browser: session.Browser,
		BrowserVersion: ua.Version,
		CreatedAt: session.CreatedAt,
		LastUsedAt: now,
		ExpiresAt: now.Add(ttl),
	}, nil
}

// TODO Review this function, it feels kinda weird

// updates session of given user
func UpdateSession(
	ctx echo.Context,
	session *SessionDTO.Full, // can be nil
	user *UserDTO.Full,
	payload *UserDTO.Payload, // from existing token claims
) (accessToken *token.SignedToken, refreshToken *token.SignedToken, err *Error.Status){
	if user == nil {
		controller.Log.Panic(
			"Invalid updateSession call",
			"user is nil (expected *UserDTO.Full)",
			nil,
		)
		return nil, nil, Error.StatusInternalError
	}
	if payload == nil {
		controller.Log.Panic(
			"Invalid updateSession call",
			"payload is nil (expected *UserDTO.Payload)",
			nil,
		)
		return nil, nil, Error.StatusInternalError
	}

	isSessionSet := session != nil

	if user.Version != payload.Version {
		payload.ID = user.ID
		payload.Login = user.Login
		payload.Roles = user.Roles
		payload.Version = user.Version
	}

	act := ActionDTO.NewUserTargeted(payload.ID, payload.ID, payload.Roles)

	if !isSessionSet {
		var err *Error.Status

		session, err = DB.Database.GetSessionByID(act, payload.SessionID)
		if err != nil {
			return nil, nil, err
		}
	}

	accessToken, refreshToken, err = token.NewAuthTokens(payload)
	if err != nil {
		return nil, nil, err
	}

	newSession, err := actualizeSession(ctx, session, config.Auth.RefreshTokenTTL())
	if err != nil {
		return nil, nil, err
	}

	// If session was specified in this function args need to ensure that this session is exists in DB.
	// If it wasn't specified then this session is queried from DB in this function, so there are no need in this check.
	if isSessionSet {
		// Check if this session exists in DB
		if _, err := DB.Database.GetSessionByID(act, newSession.ID); err != nil {
			return nil, nil, err
		}
	}

	if err := DB.Database.UpdateSession(&act.Basic, newSession.ID, newSession); err != nil {
		return nil, nil, err
	}

	if _, err := updateOrCreateLocation(act, newSession.ID, newSession.IpAddress); err != nil {
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}

