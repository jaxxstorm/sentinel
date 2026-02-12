package diff

import (
	"context"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

type PresenceDetector struct {
	now func() time.Time
}

func NewPresenceDetector() *PresenceDetector {
	return &PresenceDetector{now: time.Now}
}

func (d *PresenceDetector) Name() string { return "presence" }

func (d *PresenceDetector) Detect(_ context.Context, before, after snapshot.Snapshot) ([]event.Event, error) {
	prev := snapshot.IndexByPeerID(before)
	next := snapshot.IndexByPeerID(after)
	result := make([]event.Event, 0)

	for id, p := range next {
		old, exists := prev[id]
		if !exists {
			if p.Online {
				result = append(result, event.NewPresenceEvent(event.TypePeerOnline, id, before.Hash, after.Hash, map[string]any{"name": p.Name}, d.now()))
			}
			continue
		}
		if old.Online == p.Online {
			continue
		}
		eventType := event.TypePeerOffline
		if p.Online {
			eventType = event.TypePeerOnline
		}
		result = append(result, event.NewPresenceEvent(eventType, id, before.Hash, after.Hash, map[string]any{"name": p.Name}, d.now()))
	}
	for id, old := range prev {
		if _, exists := next[id]; exists {
			continue
		}
		if !old.Online {
			continue
		}
		result = append(result, event.NewPresenceEvent(event.TypePeerOffline, id, before.Hash, after.Hash, map[string]any{"name": old.Name, "reason": "missing_from_netmap"}, d.now()))
	}
	return result, nil
}
