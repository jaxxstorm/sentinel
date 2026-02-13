package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/state"
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
