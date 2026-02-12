package policy

import (
	"fmt"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
)

type Config struct {
	DebounceWindow    time.Duration
	SuppressionWindow time.Duration
	RateLimitPerMin   int
	BatchSize         int
}

type SuppressedEvent struct {
	Event  event.Event
	Reason string
}

type Result struct {
	Batches    [][]event.Event
	Suppressed []SuppressedEvent
}

type Engine struct {
	cfg          Config
	lastSeen     map[string]time.Time
	rateWindow   time.Time
	rateConsumed int
	now          func() time.Time
}

func NewEngine(cfg Config) *Engine {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1
	}
	return &Engine{
		cfg:      cfg,
		lastSeen: map[string]time.Time{},
		now:      time.Now,
	}
}

func (e *Engine) Apply(events []event.Event) (Result, error) {
	res := Result{}
	accepted := make([]event.Event, 0, len(events))
	now := e.now().UTC()

	for _, evt := range events {
		key := fmt.Sprintf("%s|%s", evt.EventType, evt.SubjectID)
		if t, ok := e.lastSeen[key]; ok {
			if e.cfg.DebounceWindow > 0 && now.Sub(t) < e.cfg.DebounceWindow {
				res.Suppressed = append(res.Suppressed, SuppressedEvent{Event: evt, Reason: "debounce"})
				continue
			}
			if e.cfg.SuppressionWindow > 0 && now.Sub(t) < e.cfg.SuppressionWindow {
				res.Suppressed = append(res.Suppressed, SuppressedEvent{Event: evt, Reason: "suppression"})
				continue
			}
		}

		if e.cfg.RateLimitPerMin > 0 {
			if e.rateWindow.IsZero() || now.Sub(e.rateWindow) >= time.Minute {
				e.rateWindow = now
				e.rateConsumed = 0
			}
			if e.rateConsumed >= e.cfg.RateLimitPerMin {
				res.Suppressed = append(res.Suppressed, SuppressedEvent{Event: evt, Reason: "rate_limit"})
				continue
			}
			e.rateConsumed++
		}

		e.lastSeen[key] = now
		accepted = append(accepted, evt)
	}

	for i := 0; i < len(accepted); i += e.cfg.BatchSize {
		end := i + e.cfg.BatchSize
		if end > len(accepted) {
			end = len(accepted)
		}
		res.Batches = append(res.Batches, accepted[i:end])
	}
	return res, nil
}
