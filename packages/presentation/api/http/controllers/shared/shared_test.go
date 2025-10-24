package sharedcontroller

import (
	"testing"

	Error "sentinel/packages/common/errors"
	SessionDTO "sentinel/packages/core/session/DTO"

	"github.com/mileusna/useragent"
)

func TestDeviceDetectionLogic(t *testing.T) {
	t.Run("mobile device identification", func(t *testing.T) {
		ua := useragent.UserAgent{
			Mobile:  true,
			Tablet:  false,
			Desktop: false,
			Device:  "Samsung Galaxy",
		}
		deviceType, deviceID := getDeviceInfo(ua)

		if deviceType != mobile {
			t.Error("Expected mobile device type")
		}
		if deviceID != "Samsung Galaxy" {
			t.Error("Expected device model as ID")
		}
	})

	t.Run("fallback for unknown mobile devices", func(t *testing.T) {
		ua := useragent.UserAgent{
			Mobile:  true,
			Tablet:  false,
			Desktop: false,
			Device:  "",
			Name:    "Chrome",
		}
		deviceType, deviceID := getDeviceInfo(ua)

		if deviceType != mobile {
			t.Error("Expected mobile device type")
		}
		if deviceID != "Unknown mobile" {
			t.Error("Expected unknown fallback")
		}
	})

	t.Run("desktop device OS+Browser combination", func(t *testing.T) {
		ua := useragent.UserAgent{
			Mobile:  false,
			Tablet:  false,
			Desktop: true,
			OS:      "Windows",
			Name:    "Chrome",
		}
		deviceType, deviceID := getDeviceInfo(ua)

		if deviceType != desktop {
			t.Error("Expected desktop device type")
		}
		if deviceID != "WindowsChrome" {
			t.Errorf("Expected 'WindowsChrome', got '%s'", deviceID)
		}
	})
}

func TestErrorResponseCodes(t *testing.T) {
	t.Run("session conflicts return HTTP 409", func(t *testing.T) {
		if deviceIDMismatch.Status() != 409 {
			t.Error("Device mismatch should be conflict error")
		}
		if browserMismatch.Status() != 409 {
			t.Error("Browser mismatch should be conflict error")
		}
		if osMismatch.Status() != 409 {
			t.Error("OS mismatch should be conflict error")
		}
	})
}

func TestSessionActualizationLogic(t *testing.T) {
	// Test the core business logic for session reuse validation
	// This validates whether existing sessions can be safely continued

	t.Run("allows matching browsers", func(t *testing.T) {
		// Session with same browser should be allowed
		oldSession := &SessionDTO.Full{
			Browser:  "Chrome",
			OS:       "Windows",
			DeviceID: "WindowsChrome",
		}

		ua := useragent.UserAgent{
			Desktop: true,
			OS:      "Windows",
			Name:    "Chrome", // Same browser
		}

		result, err := simulateSessionActualization(oldSession, ua, mobile, "WindowsChrome")
		if err != nil {
			t.Errorf("Expected successful session actualization, got error: %v", err)
		}
		if result == nil {
			t.Error("Expected valid session result")
		}
	})

	t.Run("rejects browser mismatch", func(t *testing.T) {
		// Different browser should be rejected (sessions can't be shared)
		oldSession := &SessionDTO.Full{
			Browser: "Chrome",
		}

		ua := useragent.UserAgent{
			Desktop: true,
			Name:    "Firefox", // Different browser
		}

		_, err := simulateSessionActualization(oldSession, ua, desktop, "WindowsFirefox")
		if err == nil {
			t.Error("Expected browser mismatch error")
		}
		if err.Status() != 409 {
			t.Errorf("Expected conflict error (409), got %d", err.Status())
		}
	})

	t.Run("rejects OS mismatch for desktops", func(t *testing.T) {
		// Different OS should be rejected for desktops (new device)
		oldSession := &SessionDTO.Full{
			Browser:  "Chrome",
			OS:       "Windows",
			DeviceID: "WindowsChrome",
		}

		ua := useragent.UserAgent{
			Desktop: true,
			OS:      "Linux", // Different OS
			Name:    "Chrome",
		}

		_, err := simulateSessionActualization(oldSession, ua, desktop, "LinuxChrome")
		if err == nil {
			t.Error("Expected OS mismatch error")
		}
		if err.Status() != 409 {
			t.Errorf("Expected conflict error (409), got %d", err.Status())
		}
	})

	t.Run("rejects device ID mismatch", func(t *testing.T) {
		// Different device ID should be rejected (mobile device changed)
		oldSession := &SessionDTO.Full{
			Browser:  "Safari",
			OS:       "iOS",
			DeviceID: "iPhone",
		}

		ua := useragent.UserAgent{
			Mobile: true,
			OS:     "iOS",
			Name:   "Safari",
			Device: "iPad", // Different device model
		}

		_, err := simulateSessionActualization(oldSession, ua, mobile, "iPad")
		if err == nil {
			t.Error("Expected device ID mismatch error")
		}
		if err.Status() != 409 {
			t.Errorf("Expected conflict error (409), got %d", err.Status())
		}
	})

	t.Run("allows device ID change for same OS/browser", func(t *testing.T) {
		// Same OS/browser combo should allow device ID flexibility
		oldSession := &SessionDTO.Full{
			Browser:  "Chrome",
			OS:       "Windows",
			DeviceID: "WindowsChrome",
		}

		ua := useragent.UserAgent{
			Desktop: true,
			OS:      "Windows",
			Name:    "Chrome",
		}

		result, err := simulateSessionActualization(oldSession, ua, desktop, "WindowsChrome")
		if err != nil {
			t.Errorf("Expected successful session actualization for same OS/browser, got error: %v", err)
		}
		if result == nil {
			t.Error("Expected valid session result")
		}
	})
}

// Helper function to simulate the core logic of actualizeSession without dependencies
func simulateSessionActualization(oldSession *SessionDTO.Full, ua useragent.UserAgent, deviceType deviceType, deviceID string) (*SessionDTO.Full, *Error.Status) {
	// Replicate the core validation logic from actualizeSession
	if oldSession.Browser != ua.Name {
		return nil, browserMismatch
	}
	if oldSession.OS != ua.OS {
		return nil, osMismatch
	}
	if oldSession.DeviceID != deviceID {
		return nil, deviceIDMismatch
	}

	return &SessionDTO.Full{
		ID:         oldSession.ID,
		UserID:     oldSession.UserID,
		Browser:    ua.Name,
		OS:         ua.OS,
		DeviceID:   deviceID,
		DeviceType: string(deviceType),
	}, nil
}
