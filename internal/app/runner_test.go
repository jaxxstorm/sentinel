package app

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/config"
	"github.com/jaxxstorm/sentinel/internal/diff"
	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/notify"
	"github.com/jaxxstorm/sentinel/internal/onboarding"
	"github.com/jaxxstorm/sentinel/internal/policy"
	"github.com/jaxxstorm/sentinel/internal/source"
	"github.com/jaxxstorm/sentinel/internal/state"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type fakeSink struct {
	sent int
}

func (f *fakeSink) Name() string { return "webhook-primary" }
func (f *fakeSink) Send(context.Context, notify.Notification) error {
	f.sent++
	return nil
}

type fakeEnrollmentManager struct {
	ensure func(context.Context) (onboarding.Status, error)
	last   onboarding.Status
	calls  int
}

func (m *fakeEnrollmentManager) EnsureEnrolled(ctx context.Context) (onboarding.Status, error) {
	m.calls++
	if m.ensure != nil {
		st, err := m.ensure(ctx)
		if err == nil {
			m.last = st
		}
		return st, err
	}
	st := onboarding.Status{State: onboarding.StateJoined, Mode: "auto"}
	m.last = st
	return st, nil
}

func (m *fakeEnrollmentManager) Probe(context.Context) (onboarding.Status, error) {
	return m.last, nil
}

func (m *fakeEnrollmentManager) LastStatus() onboarding.Status {
	return m.last
}

type fakeSource struct {
	polls      int
	assertFunc func() error
}

func (s *fakeSource) Poll(context.Context) (source.Netmap, error) {
	s.polls++
	if s.assertFunc != nil {
		if err := s.assertFunc(); err != nil {
			return source.Netmap{}, err
		}
	}
	return source.Netmap{Peers: []source.Peer{{ID: "peer1", Name: "peer1", Online: true}}}, nil
}

func TestRunOnceDryRunPipeline(t *testing.T) {
	cfg := config.Default()
	cfg.DetectorOrder = []string{"presence"}
	cfg.Detectors["presence"] = config.Detector{Enabled: true}
	cfg.Notifier.Routes = []config.RouteConfig{{EventTypes: []string{event.TypePeerOnline, event.TypePeerOffline}, Sinks: []string{"webhook-primary"}}}
	cfg.Policy.BatchSize = 10

	src := source.NewStaticSource(source.Netmap{Peers: []source.Peer{{ID: "peer1", Name: "peer1", Online: true}}})
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	nsink := &fakeSink{}
	n := notify.New(notify.Config{Routes: []notify.Route{{EventTypes: []string{event.TypePeerOnline}, Sinks: []string{"webhook-primary"}}}, IdempotencyKeyTTL: time.Hour}, store, []notify.Sink{nsink})
	r := NewRunner(
		cfg,
		src,
		diff.NewEngine([]diff.Detector{diff.NewPresenceDetector()}),
		policy.NewEngine(policy.Config{BatchSize: 10}),
		n,
		store,
		nil,
		zap.NewNop(),
		nil,
	)

	res, err := r.RunOnce(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Events) == 0 {
		t.Fatal("expected at least one event")
	}
	if nsink.sent != 0 {
		t.Fatalf("expected no sink sends in dry-run, got %d", nsink.sent)
	}
	if res.DryRunCount == 0 {
		t.Fatal("expected dry-run count > 0")
	}
}

func TestRunOnceReturnsErrorWhenEnrollmentFails(t *testing.T) {
	cfg := config.Default()
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	em := &fakeEnrollmentManager{
		ensure: func(context.Context) (onboarding.Status, error) {
			return onboarding.Status{State: onboarding.StateAuthFailed, ErrorCode: "auth_failed"}, errors.New("enrollment failed")
		},
	}
	r := NewRunner(
		cfg,
		source.NewStaticSource(source.Netmap{}),
		diff.NewEngine([]diff.Detector{diff.NewPresenceDetector()}),
		policy.NewEngine(policy.Config{BatchSize: 1}),
		notify.New(notify.Config{}, store, nil),
		store,
		nil,
		zap.NewNop(),
		em,
	)
	if _, err := r.RunOnce(context.Background(), true); err == nil {
		t.Fatal("expected enrollment error")
	}
}

func TestRunOnceGatesPollingOnEnrollment(t *testing.T) {
	cfg := config.Default()
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	em := &fakeEnrollmentManager{}
	src := &fakeSource{assertFunc: func() error {
		if em.calls == 0 {
			return errors.New("poll happened before enrollment")
		}
		return nil
	}}
	r := NewRunner(
		cfg,
		src,
		diff.NewEngine([]diff.Detector{diff.NewPresenceDetector()}),
		policy.NewEngine(policy.Config{BatchSize: 1}),
		notify.New(notify.Config{}, store, nil),
		store,
		nil,
		zap.NewNop(),
		em,
	)
	if _, err := r.RunOnce(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	if src.polls != 1 {
		t.Fatalf("expected 1 poll, got %d", src.polls)
	}
}

func TestRunOnceLogsEventJSONWithoutConfiguredSinks(t *testing.T) {
	cfg := config.Default()
	cfg.Notifier.Routes = nil
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	r := NewRunner(
		cfg,
		source.NewStaticSource(source.Netmap{Peers: []source.Peer{{ID: "peer-json", Online: true}}}),
		diff.NewEngine([]diff.Detector{diff.NewPresenceDetector()}),
		policy.NewEngine(policy.Config{BatchSize: 1}),
		notify.New(notify.Config{}, store, nil),
		store,
		nil,
		logger,
		nil,
	)
	if _, err := r.RunOnce(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range logs.All() {
		if entry.Message != "netmap event" {
			continue
		}
		for _, f := range entry.Context {
			if f.Key == "event_json" && strings.Contains(f.String, "\"event_type\"") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected netmap event log to contain JSON payload")
	}
}

func TestRunOnceEnrollmentCompleteLogsOnlyOnStatusChange(t *testing.T) {
	cfg := config.Default()
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	em := &fakeEnrollmentManager{
		ensure: func(context.Context) (onboarding.Status, error) {
			return onboarding.Status{
				State:    onboarding.StateJoined,
				Mode:     "auto",
				NodeID:   "node-1",
				Hostname: "sentinel",
			}, nil
		},
		last: onboarding.Status{
			State:    onboarding.StateJoined,
			Mode:     "auto",
			NodeID:   "node-1",
			Hostname: "sentinel",
		},
	}

	r := NewRunner(
		cfg,
		source.NewStaticSource(source.Netmap{}),
		diff.NewEngine([]diff.Detector{diff.NewPresenceDetector()}),
		policy.NewEngine(policy.Config{BatchSize: 1}),
		notify.New(notify.Config{}, store, nil),
		store,
		nil,
		logger,
		em,
	)
	if _, err := r.RunOnce(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	for _, entry := range logs.All() {
		if entry.Message == "tailscale enrollment complete" {
			t.Fatal("expected no repeated enrollment complete log when status did not change")
		}
	}
}
