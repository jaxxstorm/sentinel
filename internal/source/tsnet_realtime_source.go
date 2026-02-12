package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		if note.NetMap == nil {
			continue
		}

		netmapData, err := json.Marshal(note.NetMap)
		if err != nil {
			s.cfg.Logger.Warn("failed to marshal ipnbus netmap payload", zap.Error(err))
			continue
		}
		peers, err := decodePeersFromNetMapJSON(netmapData)
		if err != nil {
			s.cfg.Logger.Warn("failed to decode ipnbus netmap payload", zap.Error(err))
			continue
		}

		s.cfg.Logger.Info("ipnbus netmap update received", zap.Int("peer_count", len(peers)))
		return Netmap{PolledAt: time.Now().UTC(), Peers: peers}, nil
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
