package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jaxxstorm/sentinel/internal/version"
)

func TestVersionCommandReleaseMetadata(t *testing.T) {
	restore := setVersionMetadata("v1.2.3", "2026-02-13T18:00:00Z", "abc1234")
	defer restore()

	cmd := newVersionCmd(&GlobalOptions{})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error running version command: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Version: v1.2.3") {
		t.Fatalf("expected tagged version output, got %q", output)
	}
	if !strings.Contains(output, "Build Timestamp: 2026-02-13T18:00:00Z") {
		t.Fatalf("expected build timestamp output, got %q", output)
	}
	if !strings.Contains(output, "Commit Hash: abc1234") {
		t.Fatalf("expected commit hash output, got %q", output)
	}
}

func TestVersionCommandFallbackMetadata(t *testing.T) {
	restore := setVersionMetadata("main", "", "")
	defer restore()

	cmd := newVersionCmd(&GlobalOptions{})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error running version command: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Version: v0.0.0-dev") {
		t.Fatalf("expected fallback version output, got %q", output)
	}
	if !strings.Contains(output, "Build Timestamp: n/a") {
		t.Fatalf("expected fallback build timestamp output, got %q", output)
	}
	if !strings.Contains(output, "Commit Hash: n/a") {
		t.Fatalf("expected fallback commit hash output, got %q", output)
	}
}

func setVersionMetadata(tag, ts, commit string) func() {
	oldTag := version.TagName
	oldTS := version.BuildTimestamp
	oldCommit := version.CommitHash

	version.TagName = tag
	version.BuildTimestamp = ts
	version.CommitHash = commit

	return func() {
		version.TagName = oldTag
		version.BuildTimestamp = oldTS
		version.CommitHash = oldCommit
	}
}
