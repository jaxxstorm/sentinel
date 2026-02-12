package source

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"tailscale.com/client/local"
	"tailscale.com/ipn"
	"tailscale.com/tailcfg"
	"tailscale.com/tsnet"
	"tailscale.com/types/netmap"
)

type watchStep struct {
	note ipn.Notify
	err  error
}

type fakeIPNBusWatcher struct {
	mu     sync.Mutex
	steps  []watchStep
	idx    int
	closed bool
}

func (w *fakeIPNBusWatcher) Next() (ipn.Notify, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.idx >= len(w.steps) {
		return ipn.Notify{}, io.EOF
	}
	step := w.steps[w.idx]
	w.idx++
	if step.err != nil {
		return ipn.Notify{}, step.err
	}
	return step.note, nil
}

func (w *fakeIPNBusWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closed = true
	return nil
}

func (w *fakeIPNBusWatcher) wasClosed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closed
}

func notifyWithPeer(stableID, name string, online bool) ipn.Notify {
	on := online
	node := (&tailcfg.Node{
		ID:           tailcfg.NodeID(1),
		StableID:     tailcfg.StableNodeID(stableID),
		Name:         name + ".tail.test.",
		ComputedName: name,
		Online:       &on,
		User:         tailcfg.UserID(123),
	}).View()

	return ipn.Notify{
		NetMap: &netmap.NetworkMap{
			Peers: []tailcfg.NodeView{node},
		},
	}
}

func TestTSNetRealtimeSourceBootstrapsFromInitialNetmap(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	logger := logging.WithSource(zap.New(core), logging.LogSourceSentinel)

	running := ipn.Running
	watcher := &fakeIPNBusWatcher{
		steps: []watchStep{
			{note: ipn.Notify{State: &running}},
			{note: notifyWithPeer("peer-1", "peer-1", true)},
		},
	}

	var watchCalls int
	src := NewTSNetRealtimeSource(&tsnet.Server{}, RealtimeConfig{
		Logger:       logger,
		ReconnectMin: time.Millisecond,
		ReconnectMax: 5 * time.Millisecond,
		NewLocalClient: func(*tsnet.Server) (*local.Client, error) {
			return &local.Client{}, nil
		},
		NewWatcher: func(context.Context, *local.Client, ipn.NotifyWatchOpt) (IPNBusWatcher, error) {
			watchCalls++
			return watcher, nil
		},
	})

	nm, err := src.Poll(context.Background())
	if err != nil {
		t.Fatalf("poll failed: %v", err)
	}
	if watchCalls != 1 {
		t.Fatalf("expected 1 watcher subscription, got %d", watchCalls)
	}
	if len(nm.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(nm.Peers))
	}
	if nm.Peers[0].ID != "peer-1" {
		t.Fatalf("expected peer id peer-1, got %q", nm.Peers[0].ID)
	}

	entries := observed.FilterMessage("ipnbus event received").All()
	if len(entries) == 0 {
		t.Fatal("expected ipnbus event received debug log")
	}
	ctx := entries[0].ContextMap()
	for _, key := range []string{"has_state", "has_prefs", "has_netmap", "has_engine", "has_error_message"} {
		if _, ok := ctx[key]; !ok {
			t.Fatalf("expected stable field %q in ipnbus event log", key)
		}
	}

	updateLogs := observed.FilterMessage("ipnbus netmap update received").All()
	if len(updateLogs) == 0 {
		t.Fatal("expected ipnbus netmap update log")
	}
	if got, ok := updateLogs[0].ContextMap()["peer_count"]; !ok || got != int64(1) {
		t.Fatalf("expected peer_count=1, got %#v", updateLogs[0].ContextMap()["peer_count"])
	}
	if got := updateLogs[0].ContextMap()[logging.LogSourceField]; got != logging.LogSourceSentinel {
		t.Fatalf("expected log_source=%q, got %#v", logging.LogSourceSentinel, got)
	}
}

func TestTSNetRealtimeSourceReconnectsAfterWatchReadFailure(t *testing.T) {
	transient := errors.New("transient watch read failure")
	firstWatcher := &fakeIPNBusWatcher{
		steps: []watchStep{{err: transient}},
	}
	secondWatcher := &fakeIPNBusWatcher{
		steps: []watchStep{{note: notifyWithPeer("peer-2", "peer-2", false)}},
	}

	watchers := []IPNBusWatcher{firstWatcher, secondWatcher}
	var idx int
	src := NewTSNetRealtimeSource(&tsnet.Server{}, RealtimeConfig{
		Logger:       zap.NewNop(),
		ReconnectMin: time.Millisecond,
		ReconnectMax: 5 * time.Millisecond,
		NewLocalClient: func(*tsnet.Server) (*local.Client, error) {
			return &local.Client{}, nil
		},
		NewWatcher: func(context.Context, *local.Client, ipn.NotifyWatchOpt) (IPNBusWatcher, error) {
			if idx >= len(watchers) {
				return nil, errors.New("unexpected extra watcher request")
			}
			w := watchers[idx]
			idx++
			return w, nil
		},
	})

	nm, err := src.Poll(context.Background())
	if err != nil {
		t.Fatalf("poll failed: %v", err)
	}
	if len(nm.Peers) != 1 || nm.Peers[0].ID != "peer-2" {
		t.Fatalf("unexpected peers after reconnect: %#v", nm.Peers)
	}
	if idx != 2 {
		t.Fatalf("expected 2 watcher subscriptions, got %d", idx)
	}
	if !firstWatcher.wasClosed() {
		t.Fatal("expected first watcher to be closed after read failure")
	}
}

func TestTSNetRealtimeSourceStopsOnContextCancellation(t *testing.T) {
	src := NewTSNetRealtimeSource(&tsnet.Server{}, RealtimeConfig{
		Logger:       zap.NewNop(),
		ReconnectMin: time.Millisecond,
		ReconnectMax: 5 * time.Millisecond,
		NewLocalClient: func(*tsnet.Server) (*local.Client, error) {
			return &local.Client{}, nil
		},
		NewWatcher: func(context.Context, *local.Client, ipn.NotifyWatchOpt) (IPNBusWatcher, error) {
			return nil, errors.New("dial failed")
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := src.Poll(ctx)
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation error, got %v", err)
	}
}
