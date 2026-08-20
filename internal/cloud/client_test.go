package cloud

import (
	"encoding/json"
	"testing"
)

func TestDeviceTokenJSONFields(t *testing.T) {
	var token DeviceToken
	err := json.Unmarshal([]byte(`{
      "user":{"id":"user-1","email":"owner@example.com"},
      "device":{"id":"device-1","user_id":"user-1","name":"laptop","platform":"darwin","public_key":"AQI="},
      "session":{"token":"secret","expires_at":"2026-08-20T00:00:00Z"}
    }`), &token)
	if err != nil {
		t.Fatal(err)
	}
	if token.User.ID != "user-1" || token.Device.UserID != "user-1" || token.Session.ExpiresAt == "" {
		t.Fatalf("device token fields not decoded: %+v", token)
	}
}
