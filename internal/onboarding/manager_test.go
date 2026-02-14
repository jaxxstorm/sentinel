package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type fakeProvider struct {
	checkStatus func(context.Context) (ProviderStatus, error)
	start       func(context.Context) error
	waitLogin   func(context.Context) (ProviderStatus, error)
	authKey     string
}

func (f *fakeProvider) SetAuthKey(key string) { f.authKey = key }
func (f *fakeProvider) Start(ctx context.Context) error {
	if f.start != nil {
		return f.start(ctx)
	}
	return nil
}
func (f *fakeProvider) CheckStatus(ctx context.Context) (ProviderStatus, error) {
	if f.checkStatus != nil {
		return f.checkStatus(ctx)
	}
	return ProviderStatus{}, nil
}
func (f *fakeProvider) WaitForLogin(ctx context.Context) (ProviderStatus, error) {
	if f.waitLogin != nil {
		return f.waitLogin(ctx)
	}
	return ProviderStatus{Joined: true, Hostname: "joined-via-interactive"}, nil
}

func TestResolveAuthKeyPrecedence(t *testing.T) {
	value, source := ResolveAuthKey("from-flag", "from-env", "from-config")
	if value != "from-flag" || source != "flag" {
		t.Fatalf("unexpected precedence result: value=%q source=%q", value, source)
	}
	value, source = ResolveAuthKey("", "from-env", "from-config")
	if value != "from-env" || source != "env" {
		t.Fatalf("unexpected env precedence: value=%q source=%q", value, source)
	}
}

func TestEnsureEnrolledReusesExistingState(t *testing.T) {
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			return ProviderStatus{Joined: true, Hostname: "existing-node"}, nil
		},
	}
	mgr := NewManager(Config{Mode: "auto"}, provider, nil)
	st, err := mgr.EnsureEnrolled(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.State != StateJoined {
		t.Fatalf("expected joined, got %s", st.State)
	}
}

func TestEnsureEnrolledInvalidAuthKeyClassifiedNonRetryable(t *testing.T) {
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			return ProviderStatus{Joined: false}, nil
		},
		start: func(context.Context) error {
			return errors.New("invalid auth key")
		},
	}
	mgr := NewManager(Config{Mode: "auth_key", AuthKey: "tskey-auth-k123"}, provider, nil)
	st, err := mgr.EnsureEnrolled(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if st.State != StateAuthFailed {
		t.Fatalf("expected auth_failed state, got %s", st.State)
	}
	if st.ErrorClass != ErrorClassNonRetryable {
		t.Fatalf("expected non-retryable class, got %s", st.ErrorClass)
	}
}

func TestEnsureEnrolledAuthKeySuccess(t *testing.T) {
	joined := false
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			if joined {
				return ProviderStatus{Joined: true, Hostname: "joined-node"}, nil
			}
			return ProviderStatus{Joined: false}, nil
		},
		start: func(context.Context) error {
			joined = true
			return nil
		},
	}
	mgr := NewManager(Config{Mode: "auth_key", AuthKey: "tskey-auth-k123"}, provider, nil)
	st, err := mgr.EnsureEnrolled(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.State != StateJoined {
		t.Fatalf("expected joined state, got %s", st.State)
	}
}

func TestEnsureEnrolledSetsAuthKeyBeforeProbe(t *testing.T) {
	const key = "tskey-auth-k123"
	joined := false
	checks := 0
	var provider *fakeProvider
	provider = &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			checks++
			if checks == 1 && provider.authKey != key {
				t.Fatalf("expected auth key to be set before first status check, got %q", provider.authKey)
			}
			if joined {
				return ProviderStatus{Joined: true, Hostname: "joined-node"}, nil
			}
			return ProviderStatus{Joined: false}, nil
		},
		start: func(context.Context) error {
			joined = true
			return nil
		},
	}

	mgr := NewManager(Config{Mode: "auth_key", AuthKey: key}, provider, nil)
	if _, err := mgr.EnsureEnrolled(context.Background()); err != nil {
		t.Fatal(err)
	}
	if checks < 2 {
		t.Fatalf("expected at least 2 status checks, got %d", checks)
	}
}

func TestEnsureEnrolledAuthKeyWaitsForJoin(t *testing.T) {
	const key = "tskey-auth-k123"
	started := false
	postStartChecks := 0
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			if !started {
				return ProviderStatus{Joined: false}, nil
			}
			postStartChecks++
			if postStartChecks < 3 {
				return ProviderStatus{Joined: false}, nil
			}
			return ProviderStatus{Joined: true, Hostname: "joined-node"}, nil
		},
		start: func(context.Context) error {
			started = true
			return nil
		},
	}

	mgr := NewManager(Config{Mode: "auth_key", AuthKey: key}, provider, nil)
	typed := mgr.(*manager)
	typed.authKeySettleTimeout = 25 * time.Millisecond
	typed.statusPollInterval = 1 * time.Millisecond

	st, err := mgr.EnsureEnrolled(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.State != StateJoined {
		t.Fatalf("expected joined state, got %s", st.State)
	}
	if postStartChecks < 3 {
		t.Fatalf("expected repeated post-start checks, got %d", postStartChecks)
	}
}

func TestEnsureEnrolledAuthKeyPendingClassifiedRetryable(t *testing.T) {
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			return ProviderStatus{Joined: false, NeedsLogin: false}, nil
		},
		start: func(context.Context) error {
			return nil
		},
	}
	mgr := NewManager(Config{Mode: "auth_key", AuthKey: "tskey-auth-k123"}, provider, nil)
	typed := mgr.(*manager)
	typed.authKeySettleTimeout = 2 * time.Millisecond
	typed.statusPollInterval = 1 * time.Millisecond

	st, err := mgr.EnsureEnrolled(context.Background())
	if err == nil {
		t.Fatal("expected pending error")
	}
	if st.ErrorCode != "auth_key_pending" {
		t.Fatalf("expected auth_key_pending, got %s", st.ErrorCode)
	}
	if st.ErrorClass != ErrorClassRetryable {
		t.Fatalf("expected retryable class, got %s", st.ErrorClass)
	}
}

func TestEnsureEnrolledInteractiveTimeout(t *testing.T) {
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			return ProviderStatus{Joined: false, NeedsLogin: true}, nil
		},
		waitLogin: func(ctx context.Context) (ProviderStatus, error) {
			<-ctx.Done()
			return ProviderStatus{}, ctx.Err()
		},
	}
	mgr := NewManager(Config{Mode: "interactive", LoginTimeout: 5 * time.Millisecond}, provider, nil)
	st, err := mgr.EnsureEnrolled(context.Background())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if st.ErrorCode != "login_timeout" {
		t.Fatalf("expected login_timeout, got %s", st.ErrorCode)
	}
}

func TestEnsureEnrolledInteractiveCancellation(t *testing.T) {
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			return ProviderStatus{Joined: false, NeedsLogin: true}, nil
		},
		waitLogin: func(ctx context.Context) (ProviderStatus, error) {
			<-ctx.Done()
			return ProviderStatus{}, ctx.Err()
		},
	}
	mgr := NewManager(Config{Mode: "interactive", LoginTimeout: time.Second}, provider, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	st, err := mgr.EnsureEnrolled(ctx)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if st.ErrorCode != "login_canceled" {
		t.Fatalf("expected login_canceled, got %s", st.ErrorCode)
	}
}

func TestEnsureEnrolledAuthKeyFallbackToInteractive(t *testing.T) {
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			return ProviderStatus{Joined: false, NeedsLogin: true, LoginURL: "https://login.test"}, nil
		},
		start: func(context.Context) error {
			return errors.New("invalid auth key")
		},
		waitLogin: func(context.Context) (ProviderStatus, error) {
			return ProviderStatus{Joined: true, Hostname: "interactive-joined"}, nil
		},
	}
	mgr := NewManager(Config{
		Mode:                     "auth_key",
		AuthKey:                  "tskey-auth-k123",
		AllowInteractiveFallback: true,
		LoginTimeout:             time.Second,
	}, provider, nil)
	st, err := mgr.EnsureEnrolled(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.State != StateJoined {
		t.Fatalf("expected joined after fallback, got %s", st.State)
	}
}

func TestEnrollmentErrorDoesNotLeakAuthKey(t *testing.T) {
	rawKey := "tskey-auth-k-super-secret-value"
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			return ProviderStatus{Joined: false}, nil
		},
		start: func(context.Context) error {
			return errors.New("invalid auth key")
		},
	}
	mgr := NewManager(Config{Mode: "auth_key", AuthKey: rawKey}, provider, nil)
	_, err := mgr.EnsureEnrolled(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), rawKey) {
		t.Fatal("raw auth key leaked in error output")
	}
}

func TestEnsureEnrolledDoesNotRelogJoinedStatusForBackendStateChurn(t *testing.T) {
	calls := 0
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			calls++
			backendState := "Running"
			if calls%2 == 0 {
				backendState = "Running;health=ok"
			}
			return ProviderStatus{
				Joined:       true,
				NodeID:       "node-1",
				Hostname:     "sentinel",
				BackendState: backendState,
			}, nil
		},
	}
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	mgr := NewManager(Config{Mode: "auto"}, provider, logger)

	if _, err := mgr.EnsureEnrolled(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.EnsureEnrolled(context.Background()); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range logs.All() {
		if entry.Message == "tailscale onboarding status" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected one onboarding status log, got %d", count)
	}
}

func TestProbeLogOmitsEmptyErrorFields(t *testing.T) {
	provider := &fakeProvider{
		checkStatus: func(context.Context) (ProviderStatus, error) {
			return ProviderStatus{Joined: true, Hostname: "joined-node"}, nil
		},
	}
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	mgr := NewManager(Config{Mode: "auto"}, provider, logger)

	if _, err := mgr.Probe(context.Background()); err != nil {
		t.Fatal(err)
	}

	entries := logs.FilterMessage("tailscale onboarding status").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 onboarding status log, got %d", len(entries))
	}
	ctx := entries[0].ContextMap()
	if _, ok := ctx["error_code"]; ok {
		t.Fatalf("expected error_code to be omitted, got %#v", ctx["error_code"])
	}
	if _, ok := ctx["error_class"]; ok {
		t.Fatalf("expected error_class to be omitted, got %#v", ctx["error_class"])
	}
}

func TestProbeLogIncludesErrorFieldsWhenPresent(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	mgr := NewManager(Config{Mode: "auto"}, nil, logger)

	if _, err := mgr.Probe(context.Background()); err == nil {
		t.Fatal("expected provider unavailable error")
	}

	entries := logs.FilterMessage("tailscale onboarding status").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 onboarding status log, got %d", len(entries))
	}
	ctx := entries[0].ContextMap()
	if got := ctx["error_code"]; got != "provider_unavailable" {
		t.Fatalf("expected provider_unavailable error_code, got %#v", got)
	}
	if got := ctx["error_class"]; got != string(ErrorClassNonRetryable) {
		t.Fatalf("expected %q error_class, got %#v", ErrorClassNonRetryable, got)
	}
}
