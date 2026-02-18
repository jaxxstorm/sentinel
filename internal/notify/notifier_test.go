package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/state"
	"go.uber.org/zap"
)

type fakeSink struct {
	name  string
	sends int
}

func TestNotifierSendsToMultipleSinks(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sinkA := &fakeSink{name: "sink-a"}
	sinkB := &fakeSink{name: "sink-b"}
	cfg := Config{
		Routes:            []Route{{EventTypes: []string{event.TypePeerOnline}, Sinks: []string{"sink-a", "sink-b"}}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sinkA, sinkB})
	evt := event.NewPresenceEvent(event.TypePeerOnline, "peer1", "before", "after", nil, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sinkA.sends != 1 || sinkB.sends != 1 {
		t.Fatalf("expected both sinks to receive event, got sink-a=%d sink-b=%d", sinkA.sends, sinkB.sends)
	}
}

func TestStdoutSinkWritesJSON(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	sink := NewStdoutSink("stdout-debug", buf)
	n := Notification{
		Event:          event.NewPresenceEvent(event.TypePeerOnline, "peer1", "before", "after", nil, time.Now()),
		IdempotencyKey: "k1",
	}
	if err := sink.Send(context.Background(), n); err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &decoded); err != nil {
		t.Fatalf("expected valid json output, got error: %v", err)
	}
	if _, ok := decoded["event"]; !ok {
		t.Fatalf("expected event object in sink output: %#v", decoded)
	}
	if got := decoded["log_source"]; got != "sink" {
		t.Fatalf("expected log_source=sink, got %#v", got)
	}
	if got := decoded["sink"]; got != "stdout-debug" {
		t.Fatalf("expected sink=stdout-debug, got %#v", got)
	}
	if got := decoded["idempotency_key"]; got != "k1" {
		t.Fatalf("expected idempotency_key=k1, got %#v", got)
	}
}

func (s *fakeSink) Name() string { return s.name }
func (s *fakeSink) Send(context.Context, Notification) error {
	s.sends++
	return nil
}

func TestNotifierPersistsIdempotencyAcrossInstances(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "webhook-primary"}
	cfg := Config{
		Routes:            []Route{{EventTypes: []string{event.TypePeerOnline}, Sinks: []string{"webhook-primary"}}},
		IdempotencyKeyTTL: time.Hour,
	}
	n1 := New(cfg, store, []Sink{sink})
	evt := event.NewPresenceEvent(event.TypePeerOnline, "peer1", "before", "after", nil, time.Now())
	if _, err := n1.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected 1 send, got %d", sink.sends)
	}

	n2 := New(cfg, store, []Sink{sink})
	if _, err := n2.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected duplicate suppression across restart, got %d sends", sink.sends)
	}
}

func TestNotifierWildcardRouteMatchesExpandedEvents(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-all"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-all"},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewPeerEvent(event.TypePeerRoutesChanged, "peer1", "before", "after", nil, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected wildcard route send count 1, got %d", sink.sends)
	}
}

func TestNotifierExplicitRouteDoesNotMatchDifferentType(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-explicit"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{event.TypePeerOnline},
			Sinks:      []string{"sink-explicit"},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewPeerEvent(event.TypePeerRoutesChanged, "peer1", "before", "after", nil, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 0 {
		t.Fatalf("expected explicit route to skip unmatched event, got %d sends", sink.sends)
	}
}

func TestNotifierMixedWildcardAndLiteralStillMatchesAll(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-mixed"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{event.TypePeerOnline, "*"},
			Sinks:      []string{"sink-mixed"},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewPrefsEvent(event.TypePrefsRunSSHChanged, "local", "before", "after", nil, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected mixed wildcard route send count 1, got %d", sink.sends)
	}
}

func TestNotifierDeviceNameSelectorMatches(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-device-name"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-device-name"},
			Device: DeviceSelector{
				Names: []string{"nas-01"},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewPeerEvent(event.TypePeerOnline, "peer1", "before", "after", map[string]any{
		"name":   "nas-01",
		"tags":   []string{"tag:prod"},
		"owners": []string{"123"},
		"ips":    []string{"100.64.0.10"},
	}, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected device name selector to match once, got %d", sink.sends)
	}
}

func TestNotifierDeviceSelectorsRequireAllConfiguredDimensions(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-device-and"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-device-and"},
			Device: DeviceSelector{
				Names: []string{"nas-01"},
				Tags:  []string{"tag:prod"},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewPeerEvent(event.TypePeerOnline, "peer1", "before", "after", map[string]any{
		"name":   "nas-01",
		"tags":   []string{"tag:dev"},
		"owners": []string{"123"},
		"ips":    []string{"100.64.0.10"},
	}, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 0 {
		t.Fatalf("expected selector AND semantics to reject non-matching tag, got %d sends", sink.sends)
	}
}

func TestNotifierDeviceOwnerAndIPSelectorsMatch(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-device-owner-ip"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-device-owner-ip"},
			Device: DeviceSelector{
				Owners: []string{"123"},
				IPs:    []string{"100.64.0.10"},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewPeerEvent(event.TypePeerRoutesChanged, "peer1", "before", "after", map[string]any{
		"name":   "node-a",
		"tags":   []string{"tag:prod"},
		"owners": []string{"123"},
		"ips":    []string{"100.64.0.10"},
	}, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected owner+ip selectors to match once, got %d", sink.sends)
	}
}

func TestNotifierDeviceSelectorSkipsNonPeerEvent(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-device-non-peer"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-device-non-peer"},
			Device: DeviceSelector{
				Names: []string{"nas-01"},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewDaemonEvent(event.TypeDaemonStateChanged, "local", "before", "after", map[string]any{
		"before_state": "Starting",
		"after_state":  "Running",
	}, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 0 {
		t.Fatalf("expected non-peer event to skip device selector route, got %d sends", sink.sends)
	}
}

func TestNotifierIncludeFiltersMatchTagAndIP(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-filters-include"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-filters-include"},
			Filters: RouteFilters{
				Include: NotificationFilter{
					Tags: []string{"tag:prod"},
					IPs:  []string{"100.64.0.10"},
				},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewPeerEvent(event.TypePeerOnline, "peer1", "before", "after", map[string]any{
		"name": "node-a",
		"tags": []string{"tag:prod"},
		"ips":  []string{"100.64.0.10"},
	}, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected include filters to match once, got %d", sink.sends)
	}
}

func TestNotifierExcludeFilterTakesPrecedence(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-filters-exclude"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-filters-exclude"},
			Filters: RouteFilters{
				Include: NotificationFilter{
					Tags: []string{"tag:prod"},
				},
				Exclude: NotificationFilter{
					DeviceNames: []string{"*.mullvad.ts.net"},
				},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})

	mullvad := event.NewPeerEvent(event.TypePeerOnline, "peer1", "before", "after", map[string]any{
		"name": "us-slc-wg-306.mullvad.ts.net",
		"tags": []string{"tag:prod"},
		"ips":  []string{"100.64.0.20"},
	}, time.Now())
	nonMullvad := event.NewPeerEvent(event.TypePeerOnline, "peer2", "before", "after", map[string]any{
		"name": "workstation.tail123.ts.net",
		"tags": []string{"tag:prod"},
		"ips":  []string{"100.64.0.21"},
	}, time.Now())

	if _, err := n.Notify(context.Background(), []event.Event{mullvad, nonMullvad}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected only non-mullvad event to be delivered, got %d sends", sink.sends)
	}
}

func TestNotifierFilterCIDRMatchesIdentityIP(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-filters-cidr"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-filters-cidr"},
			Filters: RouteFilters{
				Include: NotificationFilter{
					IPs: []string{"100.64.0.0/24"},
				},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewPeerEvent(event.TypePeerOnline, "peer1", "before", "after", map[string]any{
		"name": "node-a",
		"ips":  []string{"100.64.0.10"},
	}, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected CIDR include filter to match once, got %d", sink.sends)
	}
}

func TestNotifierFiltersSkipNonPeerEvents(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-filters-non-peer"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-filters-non-peer"},
			Filters: RouteFilters{
				Include: NotificationFilter{
					DeviceNames: []string{"node-a"},
				},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewDaemonEvent(event.TypeDaemonStateChanged, "daemon", "before", "after", map[string]any{
		"before_state": "starting",
		"after_state":  "running",
	}, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 0 {
		t.Fatalf("expected non-peer event to skip filtered route, got %d sends", sink.sends)
	}
}

func TestNotifierIncludeEventFilterMatchesDaemonEvent(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-filters-events-include"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-filters-events-include"},
			Filters: RouteFilters{
				Include: NotificationFilter{
					Events: []string{event.TypeDaemonStateChanged},
				},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewDaemonEvent(event.TypeDaemonStateChanged, "daemon", "before", "after", map[string]any{
		"before_state": "starting",
		"after_state":  "running",
	}, time.Now())
	if _, err := n.Notify(context.Background(), []event.Event{evt}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected include event filter to match daemon event, got %d sends", sink.sends)
	}
}

func TestNotifierExcludeEventFilterSuppressesOnlyMatchingEventTypes(t *testing.T) {
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := &fakeSink{name: "sink-filters-events-exclude"}
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{"*"},
			Sinks:      []string{"sink-filters-events-exclude"},
			Filters: RouteFilters{
				Exclude: NotificationFilter{
					Events: []string{event.TypePeerOnline},
				},
			},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	online := event.NewPeerEvent(event.TypePeerOnline, "peer1", "before", "after", nil, time.Now())
	offline := event.NewPeerEvent(event.TypePeerOffline, "peer1", "before", "after", nil, time.Now().Add(time.Second))
	if _, err := n.Notify(context.Background(), []event.Event{online, offline}, false); err != nil {
		t.Fatal(err)
	}
	if sink.sends != 1 {
		t.Fatalf("expected only non-excluded event to be delivered, got %d sends", sink.sends)
	}
}

func TestNotifierRoutesEventToDiscordSink(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	sink := NewDiscordSink("discord-primary", srv.URL, zap.NewNop())
	cfg := Config{
		Routes: []Route{{
			EventTypes: []string{event.TypePeerOnline},
			Sinks:      []string{"discord-primary"},
		}},
		IdempotencyKeyTTL: time.Hour,
	}
	n := New(cfg, store, []Sink{sink})
	evt := event.NewPresenceEvent(event.TypePeerOnline, "peer1", "before", "after", nil, time.Now())

	res, err := n.Notify(context.Background(), []event.Event{evt}, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Sent != 1 {
		t.Fatalf("expected sent=1, got %d", res.Sent)
	}
	if requests != 1 {
		t.Fatalf("expected 1 discord request, got %d", requests)
	}

	res, err = n.Notify(context.Background(), []event.Event{evt}, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Suppressed != 1 {
		t.Fatalf("expected duplicate suppression for discord route, got %d", res.Suppressed)
	}
	if requests != 1 {
		t.Fatalf("expected no extra request after suppression, got %d", requests)
	}
}
