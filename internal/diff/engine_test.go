package diff

import (
	"context"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

type fakeDetector struct {
	name  string
	event string
}

func (f fakeDetector) Name() string { return f.name }
func (f fakeDetector) Detect(context.Context, snapshot.Snapshot, snapshot.Snapshot) ([]event.Event, error) {
	return []event.Event{{EventType: f.event, Timestamp: time.Now()}}, nil
}

func TestEngineOrderingAndEnablement(t *testing.T) {
	eng := NewEngine([]Detector{
		fakeDetector{name: "presence", event: "presence"},
		fakeDetector{name: "routes", event: "routes"},
	})

	events, err := eng.Diff(context.Background(), snapshot.Snapshot{}, snapshot.Snapshot{}, []string{"presence", "routes"}, map[string]bool{"presence": true, "routes": false})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "presence" {
		t.Fatalf("expected presence first, got %s", events[0].EventType)
	}
}
