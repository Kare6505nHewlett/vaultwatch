package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newTestPagerDutyServer(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewPagerDutyNotifier_EmptyKey(t *testing.T) {
	logger := zap.NewNop()
	_, err := NewPagerDutyNotifier("", logger)
	if err == nil {
		t.Fatal("expected error for empty integration key")
	}
}

func TestNewPagerDutyNotifier_NilLogger(t *testing.T) {
	_, err := NewPagerDutyNotifier("test-key", nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestPagerDutyNotifier_Send_Warning(t *testing.T) {
	server := newTestPagerDutyServer(t, http.StatusAccepted)
	defer server.Close()

	logger := zap.NewNop()
	n, err := NewPagerDutyNotifier("test-key", logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.client = server.Client()
	// Override the URL via a round-trip rewrite by pointing client at test server.
	n.client.Transport = rewriteTransport(server.URL)

	result := CheckResult{
		SecretPath: "secret/data/myapp",
		Status:     StatusWarning,
		Message:    "expires in 24h",
	}
	if err := n.Send(result); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}
}

func TestPagerDutyNotifier_Send_Non2xx(t *testing.T) {
	server := newTestPagerDutyServer(t, http.StatusInternalServerError)
	defer server.Close()

	logger := zap.NewNop()
	n, err := NewPagerDutyNotifier("test-key", logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.client = server.Client()
	n.client.Transport = rewriteTransport(server.URL)

	result := CheckResult{
		SecretPath: "secret/data/myapp",
		Status:     StatusExpired,
		Message:    "token expired",
	}
	if err := n.Send(result); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

// rewriteTransport redirects all requests to the given base URL.
type rewriteTransport string

func (rt rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Host = req.URL.Host // keep original path
	parsed, _ := http.NewRequest(req.Method, string(rt)+req.URL.Path, req.Body)
	parsed.Header = req.Header
	return http.DefaultTransport.RoundTrip(parsed)
}
