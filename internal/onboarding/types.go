package onboarding

import (
	"context"
	"strings"
	"time"
)

type State string

const (
	StateNotJoined     State = "not_joined"
	StateLoginRequired State = "login_required"
	StateJoining       State = "joining"
	StateJoined        State = "joined"
	StateAuthFailed    State = "auth_failed"
)

type ErrorClass string

const (
	ErrorClassNone         ErrorClass = "none"
	ErrorClassRetryable    ErrorClass = "retryable"
	ErrorClassNonRetryable ErrorClass = "non_retryable"
)

type Status struct {
	State            State      `json:"state"`
	Mode             string     `json:"mode,omitempty"`
	CredentialSource string     `json:"credential_source,omitempty"`
	NodeID           string     `json:"node_id,omitempty"`
	Hostname         string     `json:"hostname,omitempty"`
	LoginURL         string     `json:"login_url,omitempty"`
	BackendState     string     `json:"backend_state,omitempty"`
	ErrorCode        string     `json:"error_code,omitempty"`
	ErrorClass       ErrorClass `json:"error_class,omitempty"`
	Message          string     `json:"message,omitempty"`
	Remediation      string     `json:"remediation,omitempty"`
}

func (s Status) Joined() bool {
	return s.State == StateJoined
}

type Config struct {
	Mode                     string
	AuthKey                  string
	AuthKeySource            string
	OAuthCredentials         OAuthCredentials
	OAuthSource              string
	AllowInteractiveFallback bool
	LoginTimeout             time.Duration
}

type OAuthCredentials struct {
	ClientSecret string
	ClientID     string
	IDToken      string
	Audience     string
}

func (o OAuthCredentials) Normalize() OAuthCredentials {
	return OAuthCredentials{
		ClientSecret: strings.TrimSpace(o.ClientSecret),
		ClientID:     strings.TrimSpace(o.ClientID),
		IDToken:      strings.TrimSpace(o.IDToken),
		Audience:     strings.TrimSpace(o.Audience),
	}
}

func (o OAuthCredentials) Configured() bool {
	return strings.TrimSpace(o.ClientSecret) != ""
}

type ProviderStatus struct {
	Joined       bool
	NeedsLogin   bool
	NodeID       string
	Hostname     string
	LoginURL     string
	BackendState string
}

type Provider interface {
	CheckStatus(ctx context.Context) (ProviderStatus, error)
	SetAuthKey(key string)
	Start(ctx context.Context) error
	WaitForLogin(ctx context.Context) (ProviderStatus, error)
}

type EnrollmentManager interface {
	EnsureEnrolled(ctx context.Context) (Status, error)
	Probe(ctx context.Context) (Status, error)
	LastStatus() Status
}

type EnrollmentError struct {
	Code    string
	Class   ErrorClass
	Message string
	Cause   error
}

func (e *EnrollmentError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *EnrollmentError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func NormalizeMode(mode string) string {
	switch mode {
	case "", "auto":
		return "auto"
	case "auth_key":
		return "auth_key"
	case "oauth":
		return "oauth"
	case "interactive":
		return "interactive"
	default:
		return ""
	}
}

func ResolveAuthKey(flagValue, envValue, configValue string) (value, source string) {
	if flagValue != "" {
		return flagValue, "flag"
	}
	if envValue != "" {
		return envValue, "env"
	}
	if configValue != "" {
		return configValue, "config"
	}
	return "", "none"
}

func ResolveOAuthCredentials(envValue, configValue OAuthCredentials) (value OAuthCredentials, source string) {
	envValue = envValue.Normalize()
	configValue = configValue.Normalize()
	value = OAuthCredentials{
		ClientSecret: firstNonEmpty(envValue.ClientSecret, configValue.ClientSecret),
		ClientID:     firstNonEmpty(envValue.ClientID, configValue.ClientID),
		IDToken:      firstNonEmpty(envValue.IDToken, configValue.IDToken),
		Audience:     firstNonEmpty(envValue.Audience, configValue.Audience),
	}
	usedEnv, usedConfig := false, false
	for _, pair := range [][2]string{
		{envValue.ClientSecret, configValue.ClientSecret},
		{envValue.ClientID, configValue.ClientID},
		{envValue.IDToken, configValue.IDToken},
		{envValue.Audience, configValue.Audience},
	} {
		switch {
		case pair[0] != "":
			usedEnv = true
		case pair[1] != "":
			usedConfig = true
		}
	}
	switch {
	case usedEnv && usedConfig:
		return value, "mixed"
	case usedEnv:
		return value, "env"
	case usedConfig:
		return value, "config"
	default:
		return value, "none"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
