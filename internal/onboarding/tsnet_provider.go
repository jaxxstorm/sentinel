package onboarding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"tailscale.com/tsnet"
)

type TSNetProvider struct {
	server       *tsnet.Server
	pollInterval time.Duration
}

func NewTSNetProvider(server *tsnet.Server) *TSNetProvider {
	return &TSNetProvider{server: server, pollInterval: 1 * time.Second}
}

func (p *TSNetProvider) SetAuthKey(key string) {
	if p.server != nil {
		p.server.AuthKey = key
	}
}

func (p *TSNetProvider) Start(ctx context.Context) error {
	_, err := p.CheckStatus(ctx)
	return err
}

func (p *TSNetProvider) CheckStatus(ctx context.Context) (ProviderStatus, error) {
	if p.server == nil {
		return ProviderStatus{}, errors.New("tsnet server is nil")
	}
	client, err := p.server.LocalClient()
	if err != nil {
		return ProviderStatus{}, fmt.Errorf("create local client: %w", err)
	}
	status, err := client.Status(ctx)
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

func (p *TSNetProvider) WaitForLogin(ctx context.Context) (ProviderStatus, error) {
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

func asString(v any) string {
	s, _ := v.(string)
	return s
}
