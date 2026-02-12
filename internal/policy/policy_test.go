package policy

import (
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
)

func TestPolicyDebounceRateLimitBatching(t *testing.T) {
	engine := NewEngine(Config{
		DebounceWindow:    2 * time.Second,
		SuppressionWindow: 0,
		RateLimitPerMin:   2,
		BatchSize:         2,
	})
	now := time.Now()
	engine.now = func() time.Time { return now }

	events := []event.Event{
		{EventType: "peer.online", SubjectID: "a"},
		{EventType: "peer.online", SubjectID: "a"}, // debounce
		{EventType: "peer.offline", SubjectID: "b"},
		{EventType: "peer.offline", SubjectID: "c"}, // rate limited
	}

	res, err := engine.Apply(events)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Suppressed) != 2 {
		t.Fatalf("expected 2 suppressed events, got %d", len(res.Suppressed))
	}
	if len(res.Batches) != 1 || len(res.Batches[0]) != 2 {
		t.Fatalf("expected one batch of 2, got %#v", res.Batches)
	}
}
