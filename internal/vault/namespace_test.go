package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"
)

func newMockNamespaceServer(t *testing.T, keys []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/namespaces" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": keys},
		})
	}))
}

func newNamespaceChecker(t *testing.T, server *httptest.Server) *NamespaceChecker {
	t.Helper()
	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewNamespaceChecker(client, zaptest.NewLogger(t))
	if err != nil {
		t.Fatalf("NewNamespaceChecker: %v", err)
	}
	return checker
}

func TestNewNamespaceChecker_NilClient(t *testing.T) {
	_, err := NewNamespaceChecker(nil, zaptest.NewLogger(t))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewNamespaceChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://localhost:8200", "token")
	_, err := NewNamespaceChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestListNamespaces_Success(t *testing.T) {
	server := newMockNamespaceServer(t, []string{"ns1/", "ns2/"})
	defer server.Close()

	checker := newNamespaceChecker(t, server)
	namespaces, err := checker.ListNamespaces(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(namespaces) != 2 {
		t.Fatalf("expected 2 namespaces, got %d", len(namespaces))
	}
	if namespaces[0].Path != "ns1/" {
		t.Errorf("expected ns1/, got %s", namespaces[0].Path)
	}
}

func TestListNamespaces_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker := newNamespaceChecker(t, server)
	_, err := checker.ListNamespaces(context.Background())
	if err == nil {
		t.Fatal("expected error on server error")
	}
}
