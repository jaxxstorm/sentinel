package diff

import (
	"context"
	"testing"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

func TestPresenceDetectorTransitionEmitsEvent(t *testing.T) {
	d := NewPresenceDetector()
	before := snapshot.Snapshot{Hash: "before", Peers: []snapshot.Peer{{ID: "peer1", Online: false, Tags: []string{"tag:dev"}, Owners: []string{"7"}, IPs: []string{"100.64.0.1"}}}}
	after := snapshot.Snapshot{Hash: "after", Peers: []snapshot.Peer{{ID: "peer1", Name: "peer-one", Online: true, Tags: []string{"tag:dev"}, Owners: []string{"7"}, IPs: []string{"100.64.0.1"}}}}

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
	assertDeviceIdentityPayload(t, events[0].Payload)
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
	assertDeviceIdentityPayload(t, events[0].Payload)
}

func assertDeviceIdentityPayload(t *testing.T, payload map[string]any) {
	t.Helper()
	if payload == nil {
		t.Fatal("expected payload")
	}
	if _, ok := payload["name"]; !ok {
		t.Fatalf("expected payload to include name, got %#v", payload)
	}
	for _, key := range []string{"tags", "owners", "ips"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected payload to include %s, got %#v", key, payload)
		}
	}
}
