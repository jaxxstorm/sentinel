package app

import (
	"context"
	"errors"
	"io"
	"net/netip"
	"path/filepath"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/config"
	"github.com/jaxxstorm/sentinel/internal/diff"
	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/notify"
	"github.com/jaxxstorm/sentinel/internal/policy"
	"github.com/jaxxstorm/sentinel/internal/source"
	"github.com/jaxxstorm/sentinel/internal/state"
	"go.uber.org/zap"
	"tailscale.com/client/local"
	"tailscale.com/ipn"
	"tailscale.com/tailcfg"
	"tailscale.com/tsnet"
	"tailscale.com/types/netmap"
)

type realtimeWatchStep struct {
	note ipn.Notify
	err  error
}

type realtimeWatcher struct {
	steps []realtimeWatchStep
	idx   int
}

func (w *realtimeWatcher) Next() (ipn.Notify, error) {
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

func (w *realtimeWatcher) Close() error { return nil }

func realtimeNotify(stableID string, online bool) ipn.Notify {
	on := online
	node := (&tailcfg.Node{
		ID:           tailcfg.NodeID(1),
		StableID:     tailcfg.StableNodeID(stableID),
		Name:         stableID + ".tail.test.",
		ComputedName: stableID,
		Online:       &on,
		User:         tailcfg.UserID(7),
	}).View()
	return ipn.Notify{NetMap: &netmap.NetworkMap{Peers: []tailcfg.NodeView{node}}}
}

func realtimeNotifyWithRoutes(stableID string, online bool, routes []string) ipn.Notify {
	on := online
	primaryRoutes := make([]netip.Prefix, 0, len(routes))
	for _, raw := range routes {
		if p, err := netip.ParsePrefix(raw); err == nil {
			primaryRoutes = append(primaryRoutes, p)
		}
	}
	node := (&tailcfg.Node{
		ID:            tailcfg.NodeID(1),
		StableID:      tailcfg.StableNodeID(stableID),
		Name:          stableID + ".tail.test.",
		ComputedName:  stableID,
		Online:        &on,
		PrimaryRoutes: primaryRoutes,
		User:          tailcfg.UserID(7),
	}).View()
	return ipn.Notify{NetMap: &netmap.NetworkMap{Peers: []tailcfg.NodeView{node}}}
}

func newRealtimeRunnerForTest(t *testing.T, watcherFactory source.IPNBusWatcherFactory) (*Runner, *fakeSink) {
	t.Helper()

	cfg := config.Default()
	cfg.Source.Mode = "realtime"
	cfg.DetectorOrder = []string{"presence"}
	cfg.Detectors["presence"] = config.Detector{Enabled: true}
	cfg.Policy.BatchSize = 10

	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{}
	notifier := notify.New(notify.Config{
		Routes: []notify.Route{{
			EventTypes: []string{event.TypePeerOnline, event.TypePeerOffline},
			Sinks:      []string{sink.Name()},
		}},
		IdempotencyKeyTTL: time.Hour,
	}, store, []notify.Sink{sink})

	src := source.NewTSNetRealtimeSource(&tsnet.Server{}, source.RealtimeConfig{
		Logger:       zap.NewNop(),
		ReconnectMin: time.Millisecond,
		ReconnectMax: 5 * time.Millisecond,
		NewLocalClient: func(*tsnet.Server) (*local.Client, error) {
			return &local.Client{}, nil
		},
		NewWatcher: watcherFactory,
	})

	r := NewRunner(
		cfg,
		src,
		diff.NewEngine([]diff.Detector{diff.NewPresenceDetector()}),
		policy.NewEngine(policy.Config{BatchSize: 10}),
		notifier,
		store,
		nil,
		zap.NewNop(),
		nil,
	)
	return r, sink
}

func TestRealtimeRunnerStartupEventDeliveredToSink(t *testing.T) {
	watcher := &realtimeWatcher{
		steps: []realtimeWatchStep{
			{note: realtimeNotify("peer-realtime", true)},
		},
	}
	runner, sink := newRealtimeRunnerForTest(t, func(context.Context, *local.Client, ipn.NotifyWatchOpt) (source.IPNBusWatcher, error) {
		return watcher, nil
	})

	res, err := runner.RunOnce(context.Background(), false)
	if err != nil {
		t.Fatalf("run once failed: %v", err)
	}
	if len(res.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(res.Events))
	}
	if res.Events[0].EventType != event.TypePeerOnline {
		t.Fatalf("expected peer.online event, got %q", res.Events[0].EventType)
	}
	if sink.sent != 1 {
		t.Fatalf("expected sink send count 1, got %d", sink.sent)
	}
}

func TestRealtimeRunnerRecoversAfterWatchFailure(t *testing.T) {
	transient := errors.New("stream interrupted")
	watchers := []source.IPNBusWatcher{
		&realtimeWatcher{steps: []realtimeWatchStep{{err: transient}}},
		&realtimeWatcher{steps: []realtimeWatchStep{{note: realtimeNotify("peer-recover", true)}}},
	}
	calls := 0
	runner, sink := newRealtimeRunnerForTest(t, func(context.Context, *local.Client, ipn.NotifyWatchOpt) (source.IPNBusWatcher, error) {
		if calls >= len(watchers) {
			return nil, errors.New("unexpected watcher request")
		}
		w := watchers[calls]
		calls++
		return w, nil
	})

	res, err := runner.RunOnce(context.Background(), false)
	if err != nil {
		t.Fatalf("run once failed: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected reconnect and resubscribe (2 calls), got %d", calls)
	}
	if len(res.Events) != 1 || res.Events[0].EventType != event.TypePeerOnline {
		t.Fatalf("expected recovered peer.online event, got %#v", res.Events)
	}
	if sink.sent != 1 {
		t.Fatalf("expected sink send count 1, got %d", sink.sent)
	}
}

func TestRealtimeRunnerDeliversRepeatedTransitions(t *testing.T) {
	// Sequence: online -> offline -> online -> offline
	// Repeated transitions should be delivered as distinct events.
	watcher := &realtimeWatcher{
		steps: []realtimeWatchStep{
			{note: realtimeNotify("peer-dup", true)},
			{note: realtimeNotify("peer-dup", false)},
			{note: realtimeNotify("peer-dup", true)},
			{note: realtimeNotify("peer-dup", false)},
		},
	}
	runner, sink := newRealtimeRunnerForTest(t, func(context.Context, *local.Client, ipn.NotifyWatchOpt) (source.IPNBusWatcher, error) {
		return watcher, nil
	})

	for i := 0; i < 4; i++ {
		if _, err := runner.RunOnce(context.Background(), false); err != nil {
			t.Fatalf("run once %d failed: %v", i+1, err)
		}
	}
	if sink.sent != 4 {
		t.Fatalf("expected 4 sink sends, got %d", sink.sent)
	}
}

func TestRealtimeRunnerSkipsNoOpSnapshotUpdates(t *testing.T) {
	watcher := &realtimeWatcher{
		steps: []realtimeWatchStep{
			{note: realtimeNotify("peer-noop", true)},
			{note: realtimeNotify("peer-noop", true)},
		},
	}
	runner, sink := newRealtimeRunnerForTest(t, func(context.Context, *local.Client, ipn.NotifyWatchOpt) (source.IPNBusWatcher, error) {
		return watcher, nil
	})

	first, err := runner.RunOnce(context.Background(), false)
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	second, err := runner.RunOnce(context.Background(), false)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if len(first.Events) != 1 {
		t.Fatalf("expected first run to emit 1 event, got %d", len(first.Events))
	}
	if len(second.Events) != 0 {
		t.Fatalf("expected second run no-op, got %d events", len(second.Events))
	}
	if sink.sent != 1 {
		t.Fatalf("expected sink send count to remain 1 after no-op, got %d", sink.sent)
	}
}

func TestRealtimeRunnerWildcardRouteDeliversExpandedEvents(t *testing.T) {
	cfg := config.Default()
	cfg.Source.Mode = "realtime"
	cfg.DetectorOrder = []string{"presence", "peer_changes", "runtime"}
	cfg.Detectors["presence"] = config.Detector{Enabled: true}
	cfg.Detectors["peer_changes"] = config.Detector{Enabled: true}
	cfg.Detectors["runtime"] = config.Detector{Enabled: true}
	cfg.Policy.BatchSize = 10

	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{}
	notifier := notify.New(notify.Config{
		Routes: []notify.Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{sink.Name()},
		}},
		IdempotencyKeyTTL: time.Hour,
	}, store, []notify.Sink{sink})

	watcher := &realtimeWatcher{
		steps: []realtimeWatchStep{
			{note: realtimeNotifyWithRoutes("peer-routes", true, []string{"10.0.0.0/24"})},
			{note: realtimeNotifyWithRoutes("peer-routes", true, []string{"10.1.0.0/24"})},
		},
	}
	src := source.NewTSNetRealtimeSource(&tsnet.Server{}, source.RealtimeConfig{
		Logger:       zap.NewNop(),
		ReconnectMin: time.Millisecond,
		ReconnectMax: 5 * time.Millisecond,
		NewLocalClient: func(*tsnet.Server) (*local.Client, error) {
			return &local.Client{}, nil
		},
		NewWatcher: func(context.Context, *local.Client, ipn.NotifyWatchOpt) (source.IPNBusWatcher, error) {
			return watcher, nil
		},
	})

	runner := NewRunner(
		cfg,
		src,
		diff.NewEngine([]diff.Detector{
			diff.NewPresenceDetector(),
			diff.NewPeerChangeDetector(),
			diff.NewRuntimeDetector(),
		}),
		policy.NewEngine(policy.Config{BatchSize: 10}),
		notifier,
		store,
		nil,
		zap.NewNop(),
		nil,
	)

	first, err := runner.RunOnce(context.Background(), false)
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if len(first.Events) == 0 {
		t.Fatal("expected first run to emit events")
	}
	firstSends := sink.sent

	second, err := runner.RunOnce(context.Background(), false)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	foundRoutesChanged := false
	for _, evt := range second.Events {
		if evt.EventType == event.TypePeerRoutesChanged {
			foundRoutesChanged = true
			break
		}
	}
	if !foundRoutesChanged {
		t.Fatalf("expected %q event, got %#v", event.TypePeerRoutesChanged, second.Events)
	}
	if sink.sent <= firstSends {
		t.Fatalf("expected wildcard route to deliver second batch, sends first=%d now=%d", firstSends, sink.sent)
	}
}

func TestRealtimeRunnerPresenceDeliveryUnchangedWithExplicitRoute(t *testing.T) {
	cfg := config.Default()
	cfg.Source.Mode = "realtime"
	cfg.DetectorOrder = []string{"presence", "peer_changes"}
	cfg.Detectors["presence"] = config.Detector{Enabled: true}
	cfg.Detectors["peer_changes"] = config.Detector{Enabled: true}
	cfg.Policy.BatchSize = 10

	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{}
	notifier := notify.New(notify.Config{
		Routes: []notify.Route{{
			EventTypes: []string{event.TypePeerOnline, event.TypePeerOffline},
			Sinks:      []string{sink.Name()},
		}},
		IdempotencyKeyTTL: time.Hour,
	}, store, []notify.Sink{sink})

	watcher := &realtimeWatcher{
		steps: []realtimeWatchStep{{note: realtimeNotify("peer-explicit", true)}},
	}
	src := source.NewTSNetRealtimeSource(&tsnet.Server{}, source.RealtimeConfig{
		Logger:       zap.NewNop(),
		ReconnectMin: time.Millisecond,
		ReconnectMax: 5 * time.Millisecond,
		NewLocalClient: func(*tsnet.Server) (*local.Client, error) {
			return &local.Client{}, nil
		},
		NewWatcher: func(context.Context, *local.Client, ipn.NotifyWatchOpt) (source.IPNBusWatcher, error) {
			return watcher, nil
		},
	})

	runner := NewRunner(
		cfg,
		src,
		diff.NewEngine([]diff.Detector{
			diff.NewPresenceDetector(),
			diff.NewPeerChangeDetector(),
		}),
		policy.NewEngine(policy.Config{BatchSize: 10}),
		notifier,
		store,
		nil,
		zap.NewNop(),
		nil,
	)

	res, err := runner.RunOnce(context.Background(), false)
	if err != nil {
		t.Fatalf("run once failed: %v", err)
	}
	if len(res.Events) < 1 {
		t.Fatal("expected at least one event")
	}
	// Presence + peer.added are generated, but explicit route should only deliver peer.online.
	if sink.sent != 1 {
		t.Fatalf("expected explicit presence route to deliver one notification, got %d", sink.sent)
	}
}
