package diff

import (
	"context"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

func TestRuntimeDetectorEmitsStatePrefsAndTailnetEvents(t *testing.T) {
	d := NewRuntimeDetector()
	d.now = func() time.Time { return time.Date(2026, 2, 13, 20, 0, 0, 0, time.UTC) }

	before := snapshot.Snapshot{
		Hash:        "before",
		DaemonState: "Starting",
		Prefs: snapshot.Prefs{
			AdvertiseRoutes: []string{"10.0.0.0/24"},
			ExitNodeID:      "node-a",
			RunSSH:          false,
			ShieldsUp:       false,
		},
		Tailnet: snapshot.Tailnet{
			Domain:     "tail-a",
			TKAEnabled: false,
		},
	}
	after := snapshot.Snapshot{
		Hash:        "after",
		DaemonState: "Running",
		Prefs: snapshot.Prefs{
			AdvertiseRoutes: []string{"10.1.0.0/24"},
			ExitNodeID:      "node-b",
			RunSSH:          true,
			ShieldsUp:       true,
		},
		Tailnet: snapshot.Tailnet{
			Domain:     "tail-b",
			TKAEnabled: true,
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
		event.TypeDaemonStateChanged,
		event.TypePrefsAdvertiseRoutesChanged,
		event.TypePrefsExitNodeChanged,
		event.TypePrefsRunSSHChanged,
		event.TypePrefsShieldsUpChanged,
		event.TypeTailnetDomainChanged,
		event.TypeTailnetTKAEnabledChanged,
	}
	for _, et := range expected {
		if !got[et] {
			t.Fatalf("expected event type %q, got %#v", et, got)
		}
	}
}

func TestRuntimeDetectorSuppressesStartupBaseline(t *testing.T) {
	d := NewRuntimeDetector()
	before := snapshot.Snapshot{}
	after := snapshot.Snapshot{
		Hash:        "after",
		DaemonState: "Running",
	}

	events, err := d.Detect(context.Background(), before, after)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no startup events, got %#v", events)
	}
}
