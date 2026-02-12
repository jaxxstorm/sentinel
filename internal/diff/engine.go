package diff

import (
	"context"
	"fmt"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

type Detector interface {
	Name() string
	Detect(ctx context.Context, before, after snapshot.Snapshot) ([]event.Event, error)
}

type Engine struct {
	detectors map[string]Detector
}

func NewEngine(ds []Detector) *Engine {
	m := make(map[string]Detector, len(ds))
	for _, d := range ds {
		m[d.Name()] = d
	}
	return &Engine{detectors: m}
}

func (e *Engine) Diff(ctx context.Context, before, after snapshot.Snapshot, order []string, enabled map[string]bool) ([]event.Event, error) {
	out := make([]event.Event, 0)
	for _, name := range order {
		d, ok := e.detectors[name]
		if !ok {
			return nil, fmt.Errorf("detector %q not registered", name)
		}
		if enabled != nil {
			en, exists := enabled[name]
			if exists && !en {
				continue
			}
		}
		events, err := d.Detect(ctx, before, after)
		if err != nil {
			return nil, fmt.Errorf("detector %q failed: %w", name, err)
		}
		out = append(out, events...)
	}
	return out, nil
}
