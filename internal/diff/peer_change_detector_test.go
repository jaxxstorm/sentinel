package diff

import (
	"context"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

func TestPeerChangeDetectorEmitsMembershipAndAttributeEvents(t *testing.T) {
	d := NewPeerChangeDetector()
	d.now = func() time.Time { return time.Date(2026, 2, 13, 20, 0, 0, 0, time.UTC) }

	before := snapshot.Snapshot{
		Hash: "before",
		Peers: []snapshot.Peer{
			{
				ID:                "peer-1",
				Name:              "peer-1",
				Online:            true,
				Tags:              []string{"tag:dev"},
				Routes:            []string{"10.0.0.0/24"},
				MachineAuthorized: false,
				KeyExpiry:         "2026-02-20T00:00:00Z",
				Expired:           false,
				HostinfoHash:      "old-hash",
			},
			{ID: "peer-removed", Name: "peer-removed", Online: true},
		},
	}
	after := snapshot.Snapshot{
		Hash: "after",
		Peers: []snapshot.Peer{
			{
				ID:                "peer-1",
				Name:              "peer-1",
				Online:            true,
				Tags:              []string{"tag:prod"},
				Routes:            []string{"10.1.0.0/24"},
				MachineAuthorized: true,
				KeyExpiry:         "2026-03-01T00:00:00Z",
				Expired:           true,
				HostinfoHash:      "new-hash",
			},
			{ID: "peer-added", Name: "peer-added", Online: false},
		},
	}

	events, err := d.Detect(context.Background(), before, after)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, e := range events {
		got[e.EventType] = true
	}
	expected := []string{
		event.TypePeerAdded,
		event.TypePeerRemoved,
		event.TypePeerRoutesChanged,
		event.TypePeerTagsChanged,
		event.TypePeerMachineAuthorizedChanged,
		event.TypePeerKeyExpiryChanged,
		event.TypePeerKeyExpired,
		event.TypePeerHostinfoChanged,
	}
	for _, et := range expected {
		if !got[et] {
			t.Fatalf("expected event type %q, got %#v", et, got)
		}
	}
}

func TestPeerChangeDetectorUnchangedEmitsNone(t *testing.T) {
	d := NewPeerChangeDetector()
	before := snapshot.Snapshot{
		Hash: "before",
		Peers: []snapshot.Peer{{
			ID:                "peer-1",
			Name:              "peer-1",
			Online:            true,
			Tags:              []string{"tag:dev"},
			Routes:            []string{"10.0.0.0/24"},
			MachineAuthorized: true,
			KeyExpiry:         "2026-02-20T00:00:00Z",
			Expired:           false,
			HostinfoHash:      "hash",
		}},
	}
	after := before
	after.Hash = "after"

	events, err := d.Detect(context.Background(), before, after)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no events, got %#v", events)
	}
}
