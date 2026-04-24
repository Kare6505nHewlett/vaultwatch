package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockTokenListServer(t *testing.T, statusCode int, accessors []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "LIST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			body, _ := json.Marshal(map[string]interface{}{
				"data": map[string]interface{}{"keys": accessors},
			})
			w.Write(body)
		}
	}))
}

func newTokenLister(t *testing.T, serverURL string) *TokenLister {
	t.Helper()
	client := &Client{
		Address: serverURL,
		Token:   "test-token",
		HTTP:    &http.Client{},
	}
	lister, err := NewTokenLister(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTokenLister: %v", err)
	}
	return lister
}

func TestNewTokenLister_NilClient(t *testing.T) {
	_, err := NewTokenLister(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTokenLister_NilLogger(t *testing.T) {
	client := &Client{HTTP: &http.Client{}}
	_, err := NewTokenLister(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestListTokens_Success(t *testing.T) {
	accessors := []string{"acc-1", "acc-2", "acc-3"}
	server := newMockTokenListServer(t, http.StatusOK, accessors)
	defer server.Close()

	lister := newTokenLister(t, server.URL)
	result, err := lister.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
	if len(result.Accessors) != 3 {
		t.Errorf("expected 3 accessors, got %d", len(result.Accessors))
	}
	if result.Accessors[0] != "acc-1" {
		t.Errorf("unexpected first accessor: %s", result.Accessors[0])
	}
}

func TestListTokens_ServerError(t *testing.T) {
	server := newMockTokenListServer(t, http.StatusInternalServerError, nil)
	defer server.Close()

	lister := newTokenLister(t, server.URL)
	_, err := lister.ListTokens()
	if err == nil {
		t.Fatal("expected error on server error response")
	}
}
