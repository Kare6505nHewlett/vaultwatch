package vault

import (
	"context"
	"fmt"
	"log/slog"
)

// PolicyInfo holds metadata about a Vault policy.
type PolicyInfo struct {
	Name  string
	Rules string
}

// PolicyChecker fetches policy information from Vault.
type PolicyChecker struct {
	client *Client
	logger *slog.Logger
}

// NewPolicyChecker creates a new PolicyChecker.
func NewPolicyChecker(client *Client, logger *slog.Logger) (*PolicyChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &PolicyChecker{client: client, logger: logger}, nil
}

// GetPolicy retrieves a named policy's rules from Vault.
func (p *PolicyChecker) GetPolicy(ctx context.Context, name string) (*PolicyInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("policy name must not be empty")
	}

	path := fmt.Sprintf("/v1/sys/policy/%s", name)
	data, err := p.client.RawGet(ctx, path)
	if err != nil {
		p.logger.Error("failed to fetch policy", "policy", name, "error", err)
		return nil, fmt.Errorf("get policy %q: %w", name, err)
	}

	rules, _ := data["rules"].(string)
	p.logger.Info("fetched policy", "policy", name)
	return &PolicyInfo{Name: name, Rules: rules}, nil
}

// ListPolicies returns all policy names from Vault.
func (p *PolicyChecker) ListPolicies(ctx context.Context) ([]string, error) {
	data, err := p.client.RawGet(ctx, "/v1/sys/policy")
	if err != nil {
		p.logger.Error("failed to list policies", "error", err)
		return nil, fmt.Errorf("list policies: %w", err)
	}

	raw, ok := data["policies"]
	if !ok {
		return nil, fmt.Errorf("unexpected response: missing policies key")
	}

	slice, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected policies format")
	}

	names := make([]string, 0, len(slice))
	for _, v := range slice {
		if s, ok := v.(string); ok {
			names = append(names, s)
		}
	}

	p.logger.Info("listed policies", "count", len(names))
	return names, nil
}
