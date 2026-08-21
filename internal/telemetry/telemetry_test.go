package telemetry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestReportRegistersOneEventPerVersion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		var body event
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.InstallationID == "" || body.Version != "0.4.0" || body.Source != "test" {
			t.Fatalf("unexpected event: %+v", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := report(context.Background(), server.URL, "0.4.0", "test", server.Client()); err != nil {
		t.Fatal(err)
	}
	if err := report(context.Background(), server.URL, "0.4.0", "test", server.Client()); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("requests = %d, want 1", count)
	}
	if _, err := os.Stat(filepath.Join(home, ".ssher", "telemetry.json")); err != nil {
		t.Fatalf("telemetry state: %v", err)
	}
}

func TestDisabledValues(t *testing.T) {
	for _, value := range []string{"1", "true", "YES"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("SSHER_DISABLE_TELEMETRY", value)
			if !disabled() {
				t.Fatalf("%q should disable telemetry", value)
			}
		})
	}
}
