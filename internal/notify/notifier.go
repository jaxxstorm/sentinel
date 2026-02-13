package notify

import (
	"context"
	"errors"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/state"
)

type Route struct {
	EventTypes []string
	Severities []string
	Sinks      []string
}

type Config struct {
	Routes            []Route
	IdempotencyKeyTTL time.Duration
}

type Notification struct {
	Event          event.Event `json:"event"`
	IdempotencyKey string      `json:"idempotency_key"`
}

type Sink interface {
	Name() string
	Send(ctx context.Context, n Notification) error
}

type Result struct {
	Sent       int
	Suppressed int
	DryRun     int
}

type Notifier struct {
	cfg   Config
	store state.StateStore
	sinks map[string]Sink
}

func New(cfg Config, store state.StateStore, sinks []Sink) *Notifier {
	m := make(map[string]Sink, len(sinks))
	for _, sink := range sinks {
		m[sink.Name()] = sink
	}
	if cfg.IdempotencyKeyTTL <= 0 {
		cfg.IdempotencyKeyTTL = 24 * time.Hour
	}
	return &Notifier{cfg: cfg, store: store, sinks: m}
}

func (n *Notifier) Notify(ctx context.Context, events []event.Event, dryRun bool) (Result, error) {
	result := Result{}
	for _, evt := range events {
		routeTargets := n.targetsFor(evt)
		if len(routeTargets) == 0 {
			continue
		}
		key := event.DeriveIdempotencyKey(evt)
		seen, err := n.store.SeenIdempotencyKey(key)
		if err != nil {
			return result, err
		}
		if seen {
			result.Suppressed++
			continue
		}

		note := Notification{Event: evt, IdempotencyKey: key}
		if dryRun {
			result.DryRun += len(routeTargets)
			if err := n.store.RecordIdempotencyKey(key, n.cfg.IdempotencyKeyTTL); err != nil {
				return result, err
			}
			continue
		}

		for _, target := range routeTargets {
			sink, ok := n.sinks[target]
			if !ok {
				continue
			}
			if err := sink.Send(ctx, note); err != nil {
				return result, err
			}
			result.Sent++
		}
		if err := n.store.RecordIdempotencyKey(key, n.cfg.IdempotencyKeyTTL); err != nil {
			return result, err
		}
	}
	return result, nil
}

func (n *Notifier) targetsFor(evt event.Event) []string {
	out := []string{}
	for _, r := range n.cfg.Routes {
		if len(r.EventTypes) > 0 && !matchesEventType(r.EventTypes, evt.EventType) {
			continue
		}
		if len(r.Severities) > 0 && !contains(r.Severities, evt.Severity) {
			continue
		}
		out = append(out, r.Sinks...)
	}
	return uniq(out)
}

func matchesEventType(items []string, target string) bool {
	for _, item := range items {
		if item == "*" {
			return true
		}
		if item == target {
			return true
		}
	}
	return false
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func uniq(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

var ErrNoSinks = errors.New("no sinks configured")
