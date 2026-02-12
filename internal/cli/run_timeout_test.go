package cli

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/app"
	"github.com/jaxxstorm/sentinel/internal/config"
	"github.com/jaxxstorm/sentinel/internal/diff"
	"github.com/jaxxstorm/sentinel/internal/notify"
	"github.com/jaxxstorm/sentinel/internal/policy"
	"github.com/jaxxstorm/sentinel/internal/source"
	"github.com/jaxxstorm/sentinel/internal/state"
	"go.uber.org/zap"
)

func TestRunOnceWithTimeoutSetsDeadline(t *testing.T) {
	err := runOnceWithTimeout(context.Background(), func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected timeout context with deadline")
		}
		until := time.Until(deadline)
		if until < 29*time.Second || until > 31*time.Second {
			t.Fatalf("unexpected timeout window: %s", until)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRunOnceWithTimeoutPropagatesFunctionError(t *testing.T) {
	expected := errors.New("boom")
	err := runOnceWithTimeout(context.Background(), func(context.Context) error {
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected propagated error %v, got %v", expected, err)
	}
}

type errorSource struct {
	err error
}

func (s errorSource) Poll(context.Context) (source.Netmap, error) {
	return source.Netmap{}, s.err
}

func newRuntimeDepsForExecuteRunTest(t *testing.T, src source.NetmapSource) *runtimeDeps {
	t.Helper()
	cfg := config.Default()
	store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))
	r := app.NewRunner(
		cfg,
		src,
		diff.NewEngine([]diff.Detector{diff.NewPresenceDetector()}),
		policy.NewEngine(policy.Config{}),
		notify.New(notify.Config{}, store, nil),
		store,
		nil,
		zap.NewNop(),
		nil,
	)
	return &runtimeDeps{runner: r}
}

func TestExecuteRunTreatsContextCanceledAsGracefulInOnceMode(t *testing.T) {
	deps := newRuntimeDepsForExecuteRunTest(t, errorSource{err: context.Canceled})
	if err := executeRun(context.Background(), deps, true, false); err != nil {
		t.Fatalf("expected graceful nil error, got %v", err)
	}
}

func TestExecuteRunTreatsContextCanceledAsGracefulInLoopMode(t *testing.T) {
	deps := newRuntimeDepsForExecuteRunTest(t, errorSource{err: context.Canceled})
	if err := executeRun(context.Background(), deps, false, false); err != nil {
		t.Fatalf("expected graceful nil error, got %v", err)
	}
}

func TestExecuteRunPropagatesNonContextErrorInOnceMode(t *testing.T) {
	expected := errors.New("boom")
	deps := newRuntimeDepsForExecuteRunTest(t, errorSource{err: expected})
	err := executeRun(context.Background(), deps, true, false)
	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}
