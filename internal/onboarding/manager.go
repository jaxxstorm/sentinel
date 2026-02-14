package onboarding

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jaxxstorm/sentinel/internal/logging"
	"go.uber.org/zap"
)

type manager struct {
	cfg                  Config
	provider             Provider
	logger               *zap.Logger
	status               Status
	authKeySettleTimeout time.Duration
	statusPollInterval   time.Duration
}

func NewManager(cfg Config, provider Provider, logger *zap.Logger) EnrollmentManager {
	if cfg.LoginTimeout <= 0 {
		cfg.LoginTimeout = 5 * time.Minute
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &manager{
		cfg:                  cfg,
		provider:             provider,
		logger:               logger,
		status:               Status{State: StateNotJoined, Mode: NormalizeMode(cfg.Mode), ErrorClass: ErrorClassNone},
		authKeySettleTimeout: 10 * time.Second,
		statusPollInterval:   500 * time.Millisecond,
	}
}

func (m *manager) LastStatus() Status {
	return m.status
}

func (m *manager) Probe(ctx context.Context) (Status, error) {
	if m.provider == nil {
		st := Status{
			State:       StateAuthFailed,
			Mode:        NormalizeMode(m.cfg.Mode),
			ErrorCode:   "provider_unavailable",
			ErrorClass:  ErrorClassNonRetryable,
			Message:     "onboarding provider is unavailable",
			Remediation: "verify runtime wiring for tsnet onboarding provider",
		}
		m.setStatus(st)
		return st, &EnrollmentError{Code: st.ErrorCode, Class: st.ErrorClass, Message: st.Message}
	}
	m.primeProviderAuthKey()

	ps, err := m.provider.CheckStatus(ctx)
	if err != nil {
		st := Status{
			State:       StateNotJoined,
			Mode:        NormalizeMode(m.cfg.Mode),
			ErrorCode:   "status_check_failed",
			ErrorClass:  ErrorClassRetryable,
			Message:     "unable to check tailscale enrollment status",
			Remediation: "retry startup or verify tailscale connectivity",
		}
		m.setStatus(st)
		return st, err
	}
	st := statusFromProvider(ps, NormalizeMode(m.cfg.Mode))
	m.setStatus(st)
	return st, nil
}

func (m *manager) primeProviderAuthKey() {
	mode := NormalizeMode(m.cfg.Mode)
	if mode == "" || mode == "interactive" || mode == "oauth" {
		return
	}
	authKey := strings.TrimSpace(m.cfg.AuthKey)
	if authKey == "" {
		return
	}
	m.provider.SetAuthKey(authKey)
}

func (m *manager) EnsureEnrolled(ctx context.Context) (Status, error) {
	st, err := m.Probe(ctx)
	if err == nil && st.Joined() {
		return st, nil
	}

	mode := NormalizeMode(m.cfg.Mode)
	if mode == "" {
		return m.fail("invalid_login_mode", ErrorClassNonRetryable, "unsupported tailscale login mode", "use tailscale.login_mode=auto|auth_key|oauth|interactive", nil)
	}
	if mode == "auto" {
		if strings.TrimSpace(m.cfg.AuthKey) != "" {
			return m.tryAuthKey(ctx, true)
		}
		if m.cfg.OAuthCredentials.Configured() {
			return m.tryOAuth(ctx, true)
		}
		return m.tryInteractive(ctx)
	}
	if mode == "auth_key" {
		return m.tryAuthKey(ctx, m.cfg.AllowInteractiveFallback)
	}
	if mode == "oauth" {
		return m.tryOAuth(ctx, m.cfg.AllowInteractiveFallback)
	}
	return m.tryInteractive(ctx)
}

func (m *manager) tryAuthKey(ctx context.Context, allowFallback bool) (Status, error) {
	authKey := strings.TrimSpace(m.cfg.AuthKey)
	if authKey == "" {
		return m.fail("auth_key_missing", ErrorClassNonRetryable, "tailscale auth key is required for auth_key mode", "set --tailscale-auth-key, SENTINEL_TAILSCALE_AUTH_KEY, or tailscale.auth_key", nil)
	}
	if !looksLikeAuthKey(authKey) {
		return m.fail("auth_key_invalid_format", ErrorClassNonRetryable, "tailscale auth key format is invalid", "provide a valid tskey-prefixed auth key", nil)
	}

	m.provider.SetAuthKey(authKey)
	m.setStatus(Status{
		State:            StateJoining,
		Mode:             "auth_key",
		CredentialSource: m.cfg.AuthKeySource,
		Message:          "attempting auth key enrollment",
	})
	if err := m.provider.Start(ctx); err != nil {
		if allowFallback {
			m.logger.Warn("auth key onboarding failed, trying interactive fallback", zap.String("error_class", string(classifyError(err))))
			return m.tryInteractive(ctx)
		}
		return m.failWithMode("auth_key", "auth_key_start_failed", classifyError(err), "tailscale auth key enrollment failed", "verify key validity and tailnet policy", err)
	}
	ps, err := m.waitForAuthKeyEnrollment(ctx)
	if err != nil {
		if allowFallback {
			return m.tryInteractive(ctx)
		}
		return m.failWithMode("auth_key", "auth_key_status_failed", classifyError(err), "unable to verify auth key enrollment status", "retry and verify tailscale daemon state", err)
	}
	if ps.Joined {
		st := statusFromProvider(ps, "auth_key")
		st.CredentialSource = m.cfg.AuthKeySource
		m.setStatus(st)
		return st, nil
	}
	if !ps.NeedsLogin {
		return m.failWithMode("auth_key", "auth_key_pending", ErrorClassRetryable, "tailscale auth key enrollment is still in progress", "retry startup and verify tailscale connectivity", nil)
	}
	if allowFallback {
		return m.tryInteractive(ctx)
	}
	return m.failWithMode("auth_key", "auth_key_rejected", ErrorClassNonRetryable, "auth key enrollment did not join the tailnet", "verify key validity/expiry and tailnet ACL policy", nil)
}

func (m *manager) tryOAuth(ctx context.Context, allowFallback bool) (Status, error) {
	if !m.cfg.OAuthCredentials.Configured() {
		return m.failWithMode("oauth", "oauth_credentials_missing", ErrorClassNonRetryable, "oauth credentials are required for oauth mode", "set tsnet.client_secret and required companion fields", nil)
	}
	m.setStatus(Status{
		State:            StateJoining,
		Mode:             "oauth",
		CredentialSource: m.cfg.OAuthSource,
		Message:          "attempting oauth credential enrollment",
	})
	if err := m.provider.Start(ctx); err != nil {
		if allowFallback {
			m.logger.Warn("oauth onboarding failed, trying interactive fallback", zap.String("error_class", string(classifyError(err))))
			return m.tryInteractive(ctx)
		}
		return m.failWithMode("oauth", "oauth_start_failed", classifyError(err), "tailscale oauth enrollment failed", "verify oauth credential validity and tailnet policy", err)
	}
	ps, err := m.waitForAuthKeyEnrollment(ctx)
	if err != nil {
		if allowFallback {
			return m.tryInteractive(ctx)
		}
		return m.failWithMode("oauth", "oauth_status_failed", classifyError(err), "unable to verify oauth enrollment status", "retry and verify tailscale daemon state", err)
	}
	if ps.Joined {
		st := statusFromProvider(ps, "oauth")
		st.CredentialSource = m.cfg.OAuthSource
		m.setStatus(st)
		return st, nil
	}
	if !ps.NeedsLogin {
		return m.failWithMode("oauth", "oauth_pending", ErrorClassRetryable, "tailscale oauth enrollment is still in progress", "retry startup and verify tailscale connectivity", nil)
	}
	if allowFallback {
		return m.tryInteractive(ctx)
	}
	return m.failWithMode("oauth", "oauth_rejected", ErrorClassNonRetryable, "oauth enrollment did not join the tailnet", "verify oauth credentials and tailnet ACL policy", nil)
}

func (m *manager) waitForAuthKeyEnrollment(ctx context.Context) (ProviderStatus, error) {
	interval := m.statusPollInterval
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	timeout := m.authKeySettleTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var last ProviderStatus
	for {
		ps, err := m.provider.CheckStatus(ctx)
		if err != nil {
			return last, err
		}
		last = ps
		if ps.Joined || time.Now().After(deadline) {
			return ps, nil
		}
		select {
		case <-ctx.Done():
			return last, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (m *manager) tryInteractive(ctx context.Context) (Status, error) {
	probe, _ := m.provider.CheckStatus(ctx)
	m.setStatus(Status{
		State:        StateLoginRequired,
		Mode:         "interactive",
		LoginURL:     probe.LoginURL,
		BackendState: probe.BackendState,
		ErrorClass:   ErrorClassNone,
		Message:      "interactive login required",
		Remediation:  "complete tailscale login using the provided URL/code",
	})

	deadlineCtx, cancel := context.WithTimeout(ctx, m.cfg.LoginTimeout)
	defer cancel()

	ps, err := m.provider.WaitForLogin(deadlineCtx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return m.fail("login_timeout", ErrorClassRetryable, "interactive login timed out", "complete login and retry Sentinel startup", err)
		}
		if errors.Is(err, context.Canceled) {
			return m.fail("login_canceled", ErrorClassNonRetryable, "interactive login canceled", "rerun Sentinel and complete interactive login", err)
		}
		return m.fail("interactive_login_failed", classifyError(err), "interactive login failed", "verify network connectivity and retry login", err)
	}
	if !ps.Joined {
		return m.fail("interactive_login_incomplete", ErrorClassRetryable, "interactive login did not complete", "finish login and retry", nil)
	}
	st := statusFromProvider(ps, "interactive")
	m.setStatus(st)
	return st, nil
}

func (m *manager) fail(code string, class ErrorClass, msg string, remediation string, cause error) (Status, error) {
	return m.failWithMode(NormalizeMode(m.cfg.Mode), code, class, msg, remediation, cause)
}

func (m *manager) failWithMode(mode, code string, class ErrorClass, msg string, remediation string, cause error) (Status, error) {
	st := Status{
		State:       StateAuthFailed,
		Mode:        mode,
		ErrorCode:   code,
		ErrorClass:  class,
		Message:     msg,
		Remediation: remediation,
	}
	switch mode {
	case "auth_key":
		st.CredentialSource = m.cfg.AuthKeySource
	case "oauth":
		st.CredentialSource = m.cfg.OAuthSource
	}
	m.setStatus(st)
	return st, &EnrollmentError{Code: code, Class: class, Message: msg, Cause: cause}
}

func (m *manager) setStatus(st Status) {
	prev := m.status
	m.status = st
	if onboardingStatusLogKey(prev) == onboardingStatusLogKey(st) {
		return
	}
	fields := []zap.Field{
		zap.String("status", string(st.State)),
		zap.String("mode", st.Mode),
	}
	if st.ErrorCode != "" {
		fields = append(fields, zap.String("error_code", st.ErrorCode))
	}
	if st.ErrorClass != "" && st.ErrorClass != ErrorClassNone {
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
	if st.CredentialSource != "" && st.CredentialSource != "none" {
		fields = append(fields, zap.String("credential_source", st.CredentialSource))
	}
	if st.State == StateAuthFailed {
		if st.ErrorClass == ErrorClassRetryable {
			m.logger.Warn("tailscale onboarding status", fields...)
			return
		}
		m.logger.Error("tailscale onboarding status", fields...)
		return
	}
	m.logger.Info("tailscale onboarding status", fields...)
}

func onboardingStatusLogKey(st Status) string {
	return strings.Join([]string{
		string(st.State),
		st.Mode,
		st.ErrorCode,
		string(st.ErrorClass),
		st.NodeID,
		st.Hostname,
		st.LoginURL,
		st.CredentialSource,
	}, "|")
}

func statusFromProvider(ps ProviderStatus, mode string) Status {
	st := Status{
		Mode:         mode,
		NodeID:       ps.NodeID,
		Hostname:     ps.Hostname,
		LoginURL:     ps.LoginURL,
		BackendState: ps.BackendState,
		ErrorClass:   ErrorClassNone,
	}
	switch {
	case ps.Joined:
		st.State = StateJoined
	case ps.NeedsLogin:
		st.State = StateLoginRequired
		st.Message = "login required"
		st.Remediation = "complete interactive login to join tailnet"
	default:
		st.State = StateNotJoined
		st.Message = "node is not yet joined"
		st.Remediation = "configure auth key, oauth credentials, or enable interactive login"
	}
	return st
}

func looksLikeAuthKey(key string) bool {
	return strings.HasPrefix(strings.TrimSpace(key), "tskey-")
}

func classifyError(err error) ErrorClass {
	if err == nil {
		return ErrorClassNone
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "invalid") || strings.Contains(msg, "expired") || strings.Contains(msg, "unauthorized") {
		return ErrorClassNonRetryable
	}
	return ErrorClassRetryable
}

func RedactedKeyReference(key string) string {
	return logging.RedactAuthKey(key)
}

func EnrollmentSummary(st Status) string {
	if st.State == StateJoined {
		if st.Hostname != "" {
			return fmt.Sprintf("joined (%s)", st.Hostname)
		}
		return "joined"
	}
	if st.ErrorCode != "" {
		return fmt.Sprintf("%s (%s)", st.State, st.ErrorCode)
	}
	return string(st.State)
}
