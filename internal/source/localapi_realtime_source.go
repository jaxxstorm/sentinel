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
)

type LocalAPIRealtimeConfig struct {
	WatchMask    ipn.NotifyWatchOpt
	ReconnectMin time.Duration
	ReconnectMax time.Duration
	Logger       *zap.Logger
	NewWatcher   IPNBusWatcherFactory
}

type LocalAPIRealtimeSource struct {
	client  *local.Client
	cfg     LocalAPIRealtimeConfig
	watcher IPNBusWatcher
	cache   Netmap
	ready   bool
}

func NewLocalAPIRealtimeSource(client *local.Client, cfg LocalAPIRealtimeConfig) *LocalAPIRealtimeSource {
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
		cfg.NewWatcher = func(ctx context.Context, lc *local.Client, mask ipn.NotifyWatchOpt) (IPNBusWatcher, error) {
			return lc.WatchIPNBus(ctx, mask)
		}
	}
	return &LocalAPIRealtimeSource{client: client, cfg: cfg}
}

func (s *LocalAPIRealtimeSource) Poll(ctx context.Context) (Netmap, error) {
	if s.client == nil {
		return Netmap{}, errors.New("localapi client is required")
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

func (s *LocalAPIRealtimeSource) ensureWatcher(ctx context.Context) (IPNBusWatcher, error) {
	if s.watcher != nil {
		return s.watcher, nil
	}
	watcher, err := s.cfg.NewWatcher(ctx, s.client, s.cfg.WatchMask)
	if err != nil {
		return nil, fmt.Errorf("watch ipnbus: %w", err)
	}
	s.watcher = watcher
	s.cfg.Logger.Info("ipnbus watch connected")
	return s.watcher, nil
}

func (s *LocalAPIRealtimeSource) resetWatcher() {
	if s.watcher != nil {
		_ = s.watcher.Close()
		s.watcher = nil
	}
}
