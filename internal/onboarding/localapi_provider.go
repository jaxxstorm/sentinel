package onboarding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"tailscale.com/client/local"
)

type LocalAPIProvider struct {
	client       *local.Client
	pollInterval time.Duration
}

func NewLocalAPIProvider(client *local.Client) *LocalAPIProvider {
	return &LocalAPIProvider{client: client, pollInterval: 1 * time.Second}
}

func (p *LocalAPIProvider) SetAuthKey(string) {}

func (p *LocalAPIProvider) Start(ctx context.Context) error {
	_, err := p.CheckStatus(ctx)
	return err
}

func (p *LocalAPIProvider) CheckStatus(ctx context.Context) (ProviderStatus, error) {
	if p.client == nil {
		return ProviderStatus{}, errors.New("localapi client is nil")
	}
	status, err := p.client.Status(ctx)
	if err != nil {
		return ProviderStatus{}, fmt.Errorf("read tailscale status: %w", err)
	}

	data, err := json.Marshal(status)
	if err != nil {
		return ProviderStatus{}, err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return ProviderStatus{}, err
	}

	backend := asString(raw["BackendState"])
	authURL := asString(raw["AuthURL"])
	self, _ := raw["Self"].(map[string]any)
	nodeID := asString(self["ID"])
	if nodeID == "" {
		nodeID = asString(self["StableID"])
	}
	hostname := asString(self["HostName"])

	joined := strings.EqualFold(backend, "running")
	needsLogin := strings.Contains(strings.ToLower(backend), "needslogin") || authURL != ""

	return ProviderStatus{
		Joined:       joined,
		NeedsLogin:   needsLogin,
		NodeID:       nodeID,
		Hostname:     hostname,
		LoginURL:     authURL,
		BackendState: backend,
	}, nil
}

func (p *LocalAPIProvider) WaitForLogin(ctx context.Context) (ProviderStatus, error) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		st, err := p.CheckStatus(ctx)
		if err != nil {
			return ProviderStatus{}, err
		}
		if st.Joined {
			return st, nil
		}
		select {
		case <-ctx.Done():
			return ProviderStatus{}, ctx.Err()
		case <-ticker.C:
		}
	}
}
