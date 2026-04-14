package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func newMockLookupServer(t *testing.T, statusCode int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestLookupSecret_Success(t *testing.T) {
	createdAt := time.Now().UTC().Truncate(time.Second)
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"current_version": 2,
			"versions": map[string]interface{}{
				"2": map[string]interface{}{
					"created_time": createdAt.Format(time.RFC3339),
					"destroyed":    false,
				},
			},
		},
	}

	srv := newMockLookupServer(t, http.StatusOK, payload)
	defer srv.Close()

	logger := zaptest.NewLogger(t)
	client, err := NewClient(srv.URL, "test-token", logger)
	require.NoError(t, err)

	meta, err := client.LookupSecret(context.Background(), "secret", "myapp/db")
	require.NoError(t, err)
	assert.Equal(t, "secret/myapp/db", meta.Path)
	assert.Equal(t, 2, meta.Version)
	assert.False(t, meta.Destroyed)
	assert.Nil(t, meta.DeletedTime)
}

func TestLookupSecret_NotFound(t *testing.T) {
	srv := newMockLookupServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	logger := zaptest.NewLogger(t)
	client, err := NewClient(srv.URL, "test-token", logger)
	require.NoError(t, err)

	_, err = client.LookupSecret(context.Background(), "secret", "missing/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret not found")
}

func TestLookupSecret_ServerError(t *testing.T) {
	srv := newMockLookupServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	logger := zaptest.NewLogger(t)
	client, err := NewClient(srv.URL, "test-token", logger)
	require.NoError(t, err)

	_, err = client.LookupSecret(context.Background(), "secret", "some/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 500")
}

func TestLookupSecret_MissingVersion(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"current_version": 5,
			"versions":        map[string]interface{}{},
		},
	}

	srv := newMockLookupServer(t, http.StatusOK, payload)
	defer srv.Close()

	logger := zaptest.NewLogger(t)
	client, err := NewClient(srv.URL, "test-token", logger)
	require.NoError(t, err)

	_, err = client.LookupSecret(context.Background(), "secret", "some/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version 5 not found")
}
