package source

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"tailscale.com/tsnet"
)

type Peer struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Online   bool              `json:"online"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Netmap struct {
	PolledAt time.Time `json:"polled_at"`
	Peers    []Peer    `json:"peers"`
}

type NetmapSource interface {
	Poll(ctx context.Context) (Netmap, error)
}

type TSNetFetchFunc func(ctx context.Context, server *tsnet.Server) (Netmap, error)

type TSNetSource struct {
	server *tsnet.Server
	fetch  TSNetFetchFunc
}

func NewTSNetSource(server *tsnet.Server, fetch TSNetFetchFunc) *TSNetSource {
	return &TSNetSource{server: server, fetch: fetch}
}

func (s *TSNetSource) Poll(ctx context.Context) (Netmap, error) {
	if s.server == nil {
		return Netmap{}, errors.New("tsnet server is required")
	}
	if s.fetch == nil {
		return Netmap{}, errors.New("tsnet fetch function is not configured")
	}
	return s.fetch(ctx, s.server)
}

type StaticSource struct {
	netmap Netmap
}

func NewStaticSource(netmap Netmap) *StaticSource {
	return &StaticSource{netmap: netmap}
}

func (s *StaticSource) Poll(context.Context) (Netmap, error) {
	out := s.netmap
	out.PolledAt = time.Now().UTC()
	return out, nil
}

type SequenceSource struct {
	mu      sync.Mutex
	netmaps []Netmap
	idx     int
}

func NewSequenceSource(netmaps []Netmap) *SequenceSource {
	return &SequenceSource{netmaps: netmaps}
}

func (s *SequenceSource) Poll(context.Context) (Netmap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.netmaps) == 0 {
		return Netmap{}, errors.New("no netmaps configured")
	}
	if s.idx >= len(s.netmaps) {
		return Netmap{}, fmt.Errorf("sequence exhausted at index %d", s.idx)
	}
	out := s.netmaps[s.idx]
	s.idx++
	out.PolledAt = time.Now().UTC()
	return out, nil
}
