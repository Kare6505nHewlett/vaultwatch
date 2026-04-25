package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// OrphanTokenInfo holds metadata about an orphan token.
type OrphanTokenInfo struct {
	ID       string
	Orphan   bool
	Policies []string
	TTL      int
}

// OrphanTokenChecker checks whether a token is an orphan.
type OrphanTokenChecker struct {
	client *Client
	logger *log.Logger
}

// NewOrphanTokenChecker returns a new OrphanTokenChecker.
func NewOrphanTokenChecker(client *Client, logger *log.Logger) (*OrphanTokenChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &OrphanTokenChecker{client: client, logger: logger}, nil
}

// IsOrphanToken looks up the given token and reports whether it is an orphan.
func (c *OrphanTokenChecker) IsOrphanToken(token string) (*OrphanTokenInfo, error) {
	if token == "" {
		return nil, fmt.Errorf("token must not be empty")
	}

	req, err := http.NewRequest(http.MethodPost, c.client.Address+"/v1/auth/token/lookup", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.client.Token)

	body := fmt.Sprintf(`{"token":%q}`, token)
	req.Body = io.NopCloser(newStringReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			ID       string   `json:"id"`
			Orphan   bool     `json:"orphan"`
			Policies []string `json:"policies"`
			TTL      int      `json:"ttl"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.Printf("[orphan-token] token=%s orphan=%v ttl=%d", result.Data.ID, result.Data.Orphan, result.Data.TTL)
	return &OrphanTokenInfo{
		ID:       result.Data.ID,
		Orphan:   result.Data.Orphan,
		Policies: result.Data.Policies,
		TTL:      result.Data.TTL,
	}, nil
}

// newStringReader is a helper to wrap a string as an io.Reader.
func newStringReader(s string) io.Reader {
	return io.NopCloser(stringReader(s))
}

type stringReader string

func (s stringReader) Read(p []byte) (int, error) {
	copy(p, s)
	return len(s), io.EOF
}
