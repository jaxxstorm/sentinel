package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
	"tailscale.com/client/local"
	"tailscale.com/ipn"
	"tailscale.com/tsnet"
)

type IPNBusWatcher interface {
	Next() (ipn.Notify, error)
	Close() error
}

type IPNBusWatcherFactory func(ctx context.Context, client *local.Client, mask ipn.NotifyWatchOpt) (IPNBusWatcher, error)
type LocalClientFactory func(server *tsnet.Server) (*local.Client, error)

type RealtimeConfig struct {
	WatchMask      ipn.NotifyWatchOpt
	ReconnectMin   time.Duration
	ReconnectMax   time.Duration
	Logger         *zap.Logger
	NewWatcher     IPNBusWatcherFactory
	NewLocalClient LocalClientFactory
}

type TSNetRealtimeSource struct {
	server  *tsnet.Server
	cfg     RealtimeConfig
	watcher IPNBusWatcher
	cache   Netmap
	ready   bool
}

func NewTSNetRealtimeSource(server *tsnet.Server, cfg RealtimeConfig) *TSNetRealtimeSource {
	if cfg.WatchMask == 0 {
		cfg.WatchMask = ipn.NotifyInitialState | ipn.NotifyInitialPrefs | ipn.NotifyInitialNetMap | ipn.NotifyWatchEngineUpdates
	}
	if cfg.ReconnectMin <= 0 {
		cfg.ReconnectMin = 500 * time.Millisecond
	}
	if cfg.ReconnectMax <= 0 {
		cfg.ReconnectMax = 30 * time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	if cfg.NewWatcher == nil {
		cfg.NewWatcher = func(ctx context.Context, client *local.Client, mask ipn.NotifyWatchOpt) (IPNBusWatcher, error) {
			return client.WatchIPNBus(ctx, mask)
		}
	}
	if cfg.NewLocalClient == nil {
		cfg.NewLocalClient = func(server *tsnet.Server) (*local.Client, error) {
			return server.LocalClient()
		}
	}
	return &TSNetRealtimeSource{server: server, cfg: cfg}
}

func (s *TSNetRealtimeSource) Poll(ctx context.Context) (Netmap, error) {
	if s.server == nil {
		return Netmap{}, errors.New("tsnet server is required")
	}

	backoff := s.cfg.ReconnectMin
	for {
		watcher, err := s.ensureWatcher(ctx)
		if err != nil {
			s.cfg.Logger.Warn("ipnbus watch connect failed", zap.Error(err))
			if err := sleepWithContext(ctx, backoff); err != nil {
				return Netmap{}, err
			}
			backoff = minBackoff(backoff*2, s.cfg.ReconnectMax)
			continue
		}

		note, err := watcher.Next()
		if err != nil {
			s.cfg.Logger.Warn("ipnbus watch read failed", zap.Error(err))
			s.resetWatcher()
			if err := sleepWithContext(ctx, backoff); err != nil {
				return Netmap{}, err
			}
			backoff = minBackoff(backoff*2, s.cfg.ReconnectMax)
			continue
		}
		backoff = s.cfg.ReconnectMin

		s.cfg.Logger.Debug("ipnbus event received",
			zap.Bool("has_state", note.State != nil),
			zap.Bool("has_prefs", note.Prefs != nil),
			zap.Bool("has_netmap", note.NetMap != nil),
			zap.Bool("has_engine", note.Engine != nil),
			zap.Bool("has_error_message", note.ErrMessage != nil),
		)
		if note.ErrMessage != nil {
			s.cfg.Logger.Warn("ipnbus event contains error message", zap.String("error_message", *note.ErrMessage))
		}
		updated := false
		if note.State != nil {
			s.cache.DaemonState = note.State.String()
			updated = true
		}
		if note.Prefs != nil && note.Prefs.Valid() {
			routes := make([]string, 0, note.Prefs.AdvertiseRoutes().Len())
			for i := 0; i < note.Prefs.AdvertiseRoutes().Len(); i++ {
				routes = append(routes, note.Prefs.AdvertiseRoutes().At(i).String())
			}
			sort.Strings(routes)
			exitNodeID := ""
			if !note.Prefs.ExitNodeID().IsZero() {
				exitNodeID = string(note.Prefs.ExitNodeID())
			}
			s.cache.Prefs = Prefs{
				AdvertiseRoutes: routes,
				ExitNodeID:      exitNodeID,
				RunSSH:          note.Prefs.RunSSH(),
				ShieldsUp:       note.Prefs.ShieldsUp(),
			}
			updated = true
		}
		if note.ErrMessage != nil {
			s.cache.ErrorMessage = *note.ErrMessage
			updated = true
		}
		if note.NetMap == nil {
			if !updated || !s.ready {
				continue
			}
			out := cloneNetmap(s.cache)
			out.PolledAt = time.Now().UTC()
			return out, nil
		}

		netmapData, err := json.Marshal(note.NetMap)
		if err != nil {
			s.cfg.Logger.Warn("failed to marshal ipnbus netmap payload", zap.Error(err))
			continue
		}
		decoded, err := decodeNetMapJSON(netmapData)
		if err != nil {
			s.cfg.Logger.Warn("failed to decode ipnbus netmap payload", zap.Error(err))
			continue
		}
		if note.NetMap != nil {
			decoded.Tailnet.Domain = firstNonEmpty(decoded.Tailnet.Domain, note.NetMap.Domain)
			decoded.Tailnet.TKAEnabled = note.NetMap.TKAEnabled
		}
		s.cache.Peers = decoded.Peers
		s.cache.Tailnet = decoded.Tailnet
		s.ready = true
		updated = true

		if !updated {
			continue
		}
		out := cloneNetmap(s.cache)
		out.PolledAt = time.Now().UTC()
		s.cfg.Logger.Info("ipnbus netmap update received", zap.Int("peer_count", len(out.Peers)))
		return out, nil
	}
}

func (s *TSNetRealtimeSource) ensureWatcher(ctx context.Context) (IPNBusWatcher, error) {
	if s.watcher != nil {
		return s.watcher, nil
	}
	client, err := s.cfg.NewLocalClient(s.server)
	if err != nil {
		return nil, fmt.Errorf("create local client: %w", err)
	}
	watcher, err := s.cfg.NewWatcher(ctx, client, s.cfg.WatchMask)
	if err != nil {
		return nil, fmt.Errorf("watch ipnbus: %w", err)
	}
	s.watcher = watcher
	s.cfg.Logger.Info("ipnbus watch connected")
	return s.watcher, nil
}

func (s *TSNetRealtimeSource) resetWatcher() {
	if s.watcher != nil {
		_ = s.watcher.Close()
		s.watcher = nil
	}
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func minBackoff(a, b time.Duration) time.Duration {
	if b <= 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}

func cloneNetmap(nm Netmap) Netmap {
	out := nm
	out.Peers = make([]Peer, len(nm.Peers))
	for i, p := range nm.Peers {
		clone := p
		if len(p.Tags) > 0 {
			clone.Tags = append([]string(nil), p.Tags...)
		}
		if len(p.Routes) > 0 {
			clone.Routes = append([]string(nil), p.Routes...)
		}
		if len(p.Metadata) > 0 {
			meta := make(map[string]string, len(p.Metadata))
			for k, v := range p.Metadata {
				meta[k] = v
			}
			clone.Metadata = meta
		}
		out.Peers[i] = clone
	}
	if len(nm.Prefs.AdvertiseRoutes) > 0 {
		out.Prefs.AdvertiseRoutes = append([]string(nil), nm.Prefs.AdvertiseRoutes...)
	}
	return out
}
