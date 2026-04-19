package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type QuotaInfo struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Rate     float64 `json:"rate"`
	Interval float64 `json:"interval"`
}

type QuotaChecker struct {
	client *Client
	logger *zap.Logger
}

func NewQuotaChecker(client *Client, logger *zap.Logger) (*QuotaChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &QuotaChecker{client: client, logger: logger}, nil
}

func (q *QuotaChecker) ListQuotas(ctx context.Context) ([]QuotaInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, q.client.address+"/v1/sys/quotas/rate-limit?list=true", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", q.client.token)

	resp, err := q.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []QuotaInfo{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var quotas []QuotaInfo
	for _, key := range result.Data.Keys {
		quotas = append(quotas, QuotaInfo{Name: key, Type: "rate-limit"})
	}
	q.logger.Info("listed quotas", zap.Int("count", len(quotas)))
	return quotas, nil
}
