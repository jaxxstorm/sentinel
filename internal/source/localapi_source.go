package source

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tailscale.com/client/local"
)

type LocalClientFetchFunc func(ctx context.Context, client *local.Client) (Netmap, error)

type LocalClientSource struct {
	client *local.Client
	fetch  LocalClientFetchFunc
}

func NewLocalClientSource(client *local.Client, fetch LocalClientFetchFunc) *LocalClientSource {
	return &LocalClientSource{client: client, fetch: fetch}
}

func (s *LocalClientSource) Poll(ctx context.Context) (Netmap, error) {
	if s.client == nil {
		return Netmap{}, errors.New("localapi client is required")
	}
	if s.fetch == nil {
		return Netmap{}, errors.New("localapi fetch function is not configured")
	}
	return s.fetch(ctx, s.client)
}

// DefaultLocalClientFetch converts tailscaled local status JSON into Sentinel's netmap shape.
func DefaultLocalClientFetch(ctx context.Context, client *local.Client) (Netmap, error) {
	status, err := client.Status(ctx)
	if err != nil {
		return Netmap{}, fmt.Errorf("fetch tailscale status: %w", err)
	}

	nm, err := decodeNetmapFromStatus(status)
	if err != nil {
		return Netmap{}, fmt.Errorf("decode status: %w", err)
	}

	// Fallback to initial NetMap when /status does not expose peers.
	if len(nm.Peers) == 0 {
		netmap, err := fetchNetmapFromIPNBus(ctx, client)
		if err == nil && len(netmap.Peers) > 0 {
			nm = netmap
		}
	}

	nm.PolledAt = time.Now().UTC()
	return nm, nil
}
