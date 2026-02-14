package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/jaxxstorm/sentinel/internal/config"
	"github.com/jaxxstorm/sentinel/internal/diff"
	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/metrics"
	"github.com/jaxxstorm/sentinel/internal/notify"
	"github.com/jaxxstorm/sentinel/internal/onboarding"
	"github.com/jaxxstorm/sentinel/internal/policy"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
	"github.com/jaxxstorm/sentinel/internal/source"
	"github.com/jaxxstorm/sentinel/internal/state"
	"go.uber.org/zap"
)

type Runner struct {
	Cfg        config.Config
	Source     source.NetmapSource
	Diff       *diff.Engine
	Policy     *policy.Engine
	Notifier   *notify.Notifier
	Enrollment onboarding.EnrollmentManager
	State      state.StateStore
	Metrics    *metrics.Metrics
	Log        *zap.Logger
	Now        func() time.Time
	Sleep      func(time.Duration)
}

type CycleResult struct {
	Events          []event.Event
	SuppressedCount int
	SentCount       int
	DryRunCount     int
}

func NewRunner(cfg config.Config, src source.NetmapSource, d *diff.Engine, p *policy.Engine, n *notify.Notifier, st state.StateStore, m *metrics.Metrics, logger *zap.Logger, enrollment onboarding.EnrollmentManager) *Runner {
	return &Runner{
		Cfg:        cfg,
		Source:     src,
		Diff:       d,
		Policy:     p,
		Notifier:   n,
		Enrollment: enrollment,
		State:      st,
		Metrics:    m,
		Log:        logger,
		Now:        time.Now,
		Sleep:      time.Sleep,
	}
}

func (r *Runner) Run(ctx context.Context, once bool, dryRun bool) error {
	backoff := r.Cfg.PollBackoffMin
	if backoff <= 0 {
		backoff = 500 * time.Millisecond
	}
	realtimeMode := strings.EqualFold(strings.TrimSpace(r.Cfg.Source.Mode), "realtime")
	for {
		_, err := r.RunOnce(ctx, dryRun)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			if isRetryableEnrollmentError(err) {
				r.Log.Warn("poll cycle failed", zap.Error(err))
			} else {
				r.Log.Error("poll cycle failed", zap.Error(err))
			}
			if once {
				return err
			}
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			r.Sleep(backoff)
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			backoff = minDuration(backoff*2, r.Cfg.PollBackoffMax)
			continue
		}
		backoff = r.Cfg.PollBackoffMin
		if once {
			return nil
		}
		if realtimeMode {
			// In realtime mode, the source blocks until the next bus update.
			continue
		}
		wait := r.Cfg.PollInterval + jitter(r.Cfg.PollJitter)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

func (r *Runner) RunOnce(ctx context.Context, dryRun bool) (CycleResult, error) {
	start := r.Now()
	res := CycleResult{}
	if r.Enrollment != nil {
		previousEnrollmentStatus := r.Enrollment.LastStatus()
		enrollmentStatus, err := r.Enrollment.EnsureEnrolled(ctx)
		if err != nil {
			if isRetryableEnrollmentError(err) {
				r.Log.Warn("tailscale enrollment failed", enrollmentLogFields(enrollmentStatus)...)
			} else {
				r.Log.Error("tailscale enrollment failed", enrollmentLogFields(enrollmentStatus)...)
			}
			return res, fmt.Errorf("enrollment: %w", err)
		}
		if enrollmentStatusChanged(previousEnrollmentStatus, enrollmentStatus) {
			r.Log.Info("tailscale enrollment complete",
				zap.String("status", string(enrollmentStatus.State)),
				zap.String("mode", enrollmentStatus.Mode),
				zap.String("node_id", enrollmentStatus.NodeID),
				zap.String("hostname", enrollmentStatus.Hostname),
			)
		}
	}

	nm, err := r.Source.Poll(ctx)
	if err != nil {
		return res, fmt.Errorf("poll source: %w", err)
	}
	if r.Metrics != nil {
		r.Metrics.NetmapPollsTotal.Inc()
		r.Metrics.NetmapPollDurationSeconds.Observe(time.Since(start).Seconds())
	}

	current := snapshot.Normalize(nm, r.Now())
	previous, err := r.State.LoadSnapshot()
	if err != nil && !errors.Is(err, state.ErrNoSnapshot) && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, context.Canceled) {
		if r.Metrics != nil {
			r.Metrics.StateStoreErrorsTotal.Inc()
		}
		return res, fmt.Errorf("load snapshot: %w", err)
	}
	r.Log.Debug("netmap snapshot polled",
		zap.Int("peer_count", len(current.Peers)),
		zap.String("current_hash", current.Hash),
		zap.String("previous_hash", previous.Hash),
	)
	if previous.Hash != "" && previous.Hash == current.Hash {
		r.Log.Debug("no-op netmap update detected")
		return res, nil
	}

	enabled := map[string]bool{}
	for name, detector := range r.Cfg.Detectors {
		enabled[name] = detector.Enabled
	}
	events, err := r.Diff.Diff(ctx, previous, current, r.Cfg.DetectorOrder, enabled)
	if err != nil {
		return res, err
	}
	res.Events = events
	if len(events) > 0 {
		r.Log.Info("netmap diffs detected", zap.Int("events", len(events)))
		for _, evt := range events {
			evtJSON, _ := json.Marshal(evt)
			r.Log.Info("netmap event",
				zap.String("event_type", evt.EventType),
				zap.String("subject_id", evt.SubjectID),
				zap.String("subject_type", evt.SubjectType),
				zap.String("event_json", string(evtJSON)),
			)
		}
	} else {
		r.Log.Debug("no netmap diffs detected")
	}
	if r.Metrics != nil {
		for _, evt := range events {
			r.Metrics.DiffsDetectedTotal.WithLabelValues(evt.EventType).Inc()
			r.Metrics.EventsEmittedTotal.WithLabelValues(evt.EventType).Inc()
		}
	}

	policyResult, err := r.Policy.Apply(events)
	if err != nil {
		return res, fmt.Errorf("apply policy: %w", err)
	}
	r.Log.Debug("policy evaluation complete",
		zap.Int("events_in", len(events)),
		zap.Int("suppressed", len(policyResult.Suppressed)),
		zap.Int("batches", len(policyResult.Batches)),
	)
	for _, sup := range policyResult.Suppressed {
		if r.Metrics != nil {
			r.Metrics.NotificationsSuppressed.WithLabelValues(sup.Reason).Inc()
		}
	}
	res.SuppressedCount = len(policyResult.Suppressed)

	for _, batch := range policyResult.Batches {
		notifyResult, err := r.Notifier.Notify(ctx, batch, dryRun)
		if err != nil {
			return res, fmt.Errorf("notify: %w", err)
		}
		res.SentCount += notifyResult.Sent
		res.DryRunCount += notifyResult.DryRun
		for i := 0; i < notifyResult.Sent; i++ {
			if r.Metrics != nil {
				r.Metrics.NotificationsSentTotal.WithLabelValues("webhook").Inc()
			}
		}
	}

	if err := r.State.SaveSnapshot(current); err != nil {
		if r.Metrics != nil {
			r.Metrics.StateStoreErrorsTotal.Inc()
		}
		return res, fmt.Errorf("save snapshot: %w", err)
	}

	return res, nil
}

func jitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(max)))
}

func minDuration(a, b time.Duration) time.Duration {
	if b <= 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}

func enrollmentStatusChanged(before, after onboarding.Status) bool {
	return before.State != after.State ||
		before.Mode != after.Mode ||
		before.ErrorCode != after.ErrorCode ||
		before.ErrorClass != after.ErrorClass ||
		before.NodeID != after.NodeID ||
		before.Hostname != after.Hostname ||
		before.LoginURL != after.LoginURL
}

func enrollmentLogFields(st onboarding.Status) []zap.Field {
	fields := []zap.Field{
		zap.String("status", string(st.State)),
		zap.String("mode", st.Mode),
	}
	if st.ErrorCode != "" {
		fields = append(fields, zap.String("error_code", st.ErrorCode))
	}
	if st.ErrorClass != "" && st.ErrorClass != onboarding.ErrorClassNone {
		fields = append(fields, zap.String("error_class", string(st.ErrorClass)))
	}
	if st.NodeID != "" {
		fields = append(fields, zap.String("node_id", st.NodeID))
	}
	if st.Hostname != "" {
		fields = append(fields, zap.String("hostname", st.Hostname))
	}
	if st.LoginURL != "" {
		fields = append(fields, zap.String("login_url", st.LoginURL))
	}
	return fields
}

func isRetryableEnrollmentError(err error) bool {
	var enrollmentErr *onboarding.EnrollmentError
	if !errors.As(err, &enrollmentErr) {
		return false
	}
	return enrollmentErr.Class == onboarding.ErrorClassRetryable
}
