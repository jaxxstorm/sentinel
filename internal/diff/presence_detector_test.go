package diff

import (
	"context"
	"testing"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

func TestPresenceDetectorTransitionEmitsEvent(t *testing.T) {
	d := NewPresenceDetector()
	before := snapshot.Snapshot{Hash: "before", Peers: []snapshot.Peer{{ID: "peer1", Online: false}}}
	after := snapshot.Snapshot{Hash: "after", Peers: []snapshot.Peer{{ID: "peer1", Online: true}}}

	events, err := d.Detect(context.Background(), before, after)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != event.TypePeerOnline {
		t.Fatalf("expected %s, got %s", event.TypePeerOnline, events[0].EventType)
	}
}

func TestPresenceDetectorUnchangedEmitsNone(t *testing.T) {
	d := NewPresenceDetector()
	before := snapshot.Snapshot{Hash: "before", Peers: []snapshot.Peer{{ID: "peer1", Online: true}}}
	after := snapshot.Snapshot{Hash: "after", Peers: []snapshot.Peer{{ID: "peer1", Online: true}}}

	events, err := d.Detect(context.Background(), before, after)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestPresenceDetectorPeerMissingEmitsOffline(t *testing.T) {
	d := NewPresenceDetector()
	before := snapshot.Snapshot{Hash: "before", Peers: []snapshot.Peer{{ID: "peer1", Name: "peer1", Online: true}}}
	after := snapshot.Snapshot{Hash: "after", Peers: []snapshot.Peer{}}

	events, err := d.Detect(context.Background(), before, after)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != event.TypePeerOffline {
		t.Fatalf("expected %s, got %s", event.TypePeerOffline, events[0].EventType)
	}
}
