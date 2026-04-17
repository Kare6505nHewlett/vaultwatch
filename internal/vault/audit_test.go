package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockAuditServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func newAuditChecker(t *testing.T, serverURL string) *AuditChecker {
	t.Helper()
	client, err := NewClient(serverURL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewAuditChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewAuditChecker: %v", err)
	}
	return checker
}

func TestNewAuditChecker_NilClient(t *testing.T) {
	_, err := NewAuditChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewAuditChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://localhost:8200", "tok")
	_, err := NewAuditChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestListAuditDevices_Success(t *testing.T) {
	body := map[string]interface{}{
		"file/": map[string]string{"type": "file", "description": "file audit"},
		"syslog/": map[string]string{"type": "syslog", "description": "syslog audit"},
	}
	srv := newMockAuditServer(t, http.StatusOK, body)
	defer srv.Close()

	checker := newAuditChecker(t, srv.URL)
	devices, err := checker.ListAuditDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
	for _, d := range devices {
		if !d.Enabled {
			t.Errorf("expected device %s to be enabled", d.Path)
		}
	}
}

func TestListAuditDevices_ServerError(t *testing.T) {
	srv := newMockAuditServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	checker := newAuditChecker(t, srv.URL)
	_, err := checker.ListAuditDevices()
	if err == nil {
		t.Fatal("expected error on server error")
	}
}
