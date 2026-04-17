package vault

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newMockPolicyServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/policy/read-only":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name":  "read-only",
				"rules": `path "secret/*" { capabilities = ["read"] }`,
			})
		case "/v1/sys/policy":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"policies": []interface{}{"default", "read-only", "root"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newPolicyChecker(t *testing.T, srv *httptest.Server) *PolicyChecker {
	t.Helper()
	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	pc, err := NewPolicyChecker(client, logger)
	if err != nil {
		t.Fatalf("NewPolicyChecker: %v", err)
	}
	return pc
}

func TestNewPolicyChecker_NilClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := NewPolicyChecker(nil, logger)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewPolicyChecker_NilLogger(t *testing.T) {
	srv := newMockPolicyServer(t)
	defer srv.Close()
	client, _ := NewClient(srv.URL, "token")
	_, err := NewPolicyChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetPolicy_Success(t *testing.T) {
	srv := newMockPolicyServer(t)
	defer srv.Close()
	pc := newPolicyChecker(t, srv)

	info, err := pc.GetPolicy(context.Background(), "read-only")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "read-only" {
		t.Errorf("expected name read-only, got %s", info.Name)
	}
	if info.Rules == "" {
		t.Error("expected non-empty rules")
	}
}

func TestGetPolicy_EmptyName(t *testing.T) {
	srv := newMockPolicyServer(t)
	defer srv.Close()
	pc := newPolicyChecker(t, srv)

	_, err := pc.GetPolicy(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty policy name")
	}
}

func TestListPolicies_Success(t *testing.T) {
	srv := newMockPolicyServer(t)
	defer srv.Close()
	pc := newPolicyChecker(t, srv)

	names, err := pc.ListPolicies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("expected 3 policies, got %d", len(names))
	}
}
