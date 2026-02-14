package version

import "testing"

func TestResolveTaggedMetadata(t *testing.T) {
	got := resolve("v1.2.3", "2026-02-13T18:00:00Z", "abc1234")

	if got.Version != "v1.2.3" {
		t.Fatalf("expected version v1.2.3, got %q", got.Version)
	}
	if got.BuildTimestamp != "2026-02-13T18:00:00Z" {
		t.Fatalf("expected build timestamp to be preserved, got %q", got.BuildTimestamp)
	}
	if got.CommitHash != "abc1234" {
		t.Fatalf("expected commit hash to be preserved, got %q", got.CommitHash)
	}
}

func TestResolveUsesFallbackForUntaggedBuild(t *testing.T) {
	got := resolve(defaultTagName, "", "")

	if got.Version != "v0.0.0-dev" {
		t.Fatalf("expected fallback version v0.0.0-dev, got %q", got.Version)
	}
	if got.BuildTimestamp != defaultBuildTimestamp {
		t.Fatalf("expected fallback build timestamp %q, got %q", defaultBuildTimestamp, got.BuildTimestamp)
	}
	if got.CommitHash != defaultCommitHash {
		t.Fatalf("expected fallback commit hash %q, got %q", defaultCommitHash, got.CommitHash)
	}
}

func TestResolveUsesFallbackForInvalidTag(t *testing.T) {
	got := resolve("feature-branch", "2026-02-13T18:00:00Z", "abc1234")

	if got.Version != "v0.0.0-dev" {
		t.Fatalf("expected fallback version v0.0.0-dev, got %q", got.Version)
	}
}
