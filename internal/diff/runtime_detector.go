package diff

import (
	"context"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

const localSubjectID = "local"

type RuntimeDetector struct {
	now func() time.Time
}

func NewRuntimeDetector() *RuntimeDetector {
	return &RuntimeDetector{now: time.Now}
}

func (d *RuntimeDetector) Name() string { return "runtime" }

func (d *RuntimeDetector) Detect(_ context.Context, before, after snapshot.Snapshot) ([]event.Event, error) {
	// No baseline means this is startup initialization rather than a runtime transition.
	if before.Hash == "" {
		return nil, nil
	}

	out := make([]event.Event, 0)
	if before.DaemonState != "" && after.DaemonState != "" && before.DaemonState != after.DaemonState {
		out = append(out, event.NewDaemonEvent(
			event.TypeDaemonStateChanged,
			localSubjectID,
			before.Hash,
			after.Hash,
			map[string]any{
				"before_state": before.DaemonState,
				"after_state":  after.DaemonState,
			},
			d.now(),
		))
	}
	if !stringSliceEqual(before.Prefs.AdvertiseRoutes, after.Prefs.AdvertiseRoutes) {
		out = append(out, event.NewPrefsEvent(
			event.TypePrefsAdvertiseRoutesChanged,
			localSubjectID,
			before.Hash,
			after.Hash,
			map[string]any{
				"before_routes": before.Prefs.AdvertiseRoutes,
				"after_routes":  after.Prefs.AdvertiseRoutes,
			},
			d.now(),
		))
	}
	if before.Prefs.ExitNodeID != after.Prefs.ExitNodeID {
		out = append(out, event.NewPrefsEvent(
			event.TypePrefsExitNodeChanged,
			localSubjectID,
			before.Hash,
			after.Hash,
			map[string]any{
				"before_exit_node_id": before.Prefs.ExitNodeID,
				"after_exit_node_id":  after.Prefs.ExitNodeID,
			},
			d.now(),
		))
	}
	if before.Prefs.RunSSH != after.Prefs.RunSSH {
		out = append(out, event.NewPrefsEvent(
			event.TypePrefsRunSSHChanged,
			localSubjectID,
			before.Hash,
			after.Hash,
			map[string]any{
				"before_run_ssh": before.Prefs.RunSSH,
				"after_run_ssh":  after.Prefs.RunSSH,
			},
			d.now(),
		))
	}
	if before.Prefs.ShieldsUp != after.Prefs.ShieldsUp {
		out = append(out, event.NewPrefsEvent(
			event.TypePrefsShieldsUpChanged,
			localSubjectID,
			before.Hash,
			after.Hash,
			map[string]any{
				"before_shields_up": before.Prefs.ShieldsUp,
				"after_shields_up":  after.Prefs.ShieldsUp,
			},
			d.now(),
		))
	}
	if before.Tailnet.Domain != after.Tailnet.Domain {
		out = append(out, event.NewTailnetEvent(
			event.TypeTailnetDomainChanged,
			tailnetSubject(after.Tailnet.Domain),
			before.Hash,
			after.Hash,
			map[string]any{
				"before_domain": before.Tailnet.Domain,
				"after_domain":  after.Tailnet.Domain,
			},
			d.now(),
		))
	}
	if before.Tailnet.TKAEnabled != after.Tailnet.TKAEnabled {
		out = append(out, event.NewTailnetEvent(
			event.TypeTailnetTKAEnabledChanged,
			tailnetSubject(after.Tailnet.Domain),
			before.Hash,
			after.Hash,
			map[string]any{
				"before_tka_enabled": before.Tailnet.TKAEnabled,
				"after_tka_enabled":  after.Tailnet.TKAEnabled,
			},
			d.now(),
		))
	}
	return out, nil
}

func tailnetSubject(domain string) string {
	if domain == "" {
		return localSubjectID
	}
	return domain
}
