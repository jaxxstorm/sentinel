package cli

import (
	"strings"
	"testing"

	"github.com/jaxxstorm/sentinel/internal/onboarding"
)

func TestEnrollmentStatusLinesIncludeRemediation(t *testing.T) {
	lines := enrollmentStatusLines(onboarding.Status{
		State:       onboarding.StateAuthFailed,
		Mode:        "auth_key",
		ErrorCode:   "auth_key_rejected",
		Remediation: "verify auth key",
	})
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "tailscale_status=auth_failed") {
		t.Fatalf("missing state line: %s", joined)
	}
	if !strings.Contains(joined, "tailscale_remediation=verify auth key") {
		t.Fatalf("missing remediation line: %s", joined)
	}
}

func TestRenderStatusSummaryForJoined(t *testing.T) {
	summary := renderStatusSummary(onboarding.Status{State: onboarding.StateJoined, Mode: "auto", Hostname: "sentinel-1"})
	if !strings.Contains(summary, "joined") || !strings.Contains(summary, "sentinel-1") {
		t.Fatalf("unexpected summary: %s", summary)
	}
}

func TestEnrollmentStatusLinesIncludeCredentialSource(t *testing.T) {
	lines := enrollmentStatusLines(onboarding.Status{
		State:            onboarding.StateJoining,
		Mode:             "oauth",
		CredentialSource: "env",
	})
	out := strings.Join(lines, "\n")
	if !strings.Contains(out, "tailscale_credential_source=env") {
		t.Fatalf("missing credential source line: %s", out)
	}
}
