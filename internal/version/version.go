package version

import (
	"strings"

	verslib "github.com/jaxxstorm/vers"
)

const (
	defaultTagName        = "main"
	defaultBuildTimestamp = "n/a"
	defaultCommitHash     = "n/a"
)

var (
	// TagName is injected at build time by release tooling.
	TagName = defaultTagName
	// BuildTimestamp is injected at build time by release tooling.
	BuildTimestamp = defaultBuildTimestamp
	// CommitHash is injected at build time by release tooling.
	CommitHash = defaultCommitHash
)

// Metadata is the runtime version payload consumed by CLI commands and logs.
type Metadata struct {
	Version        string
	BuildTimestamp string
	CommitHash     string
}

// Current resolves runtime metadata from build-injected fields and vers fallbacks.
func Current() Metadata {
	return resolve(TagName, BuildTimestamp, CommitHash)
}

func resolve(tagName, buildTimestamp, commitHash string) Metadata {
	version := resolveVersion(tagName)
	return Metadata{
		Version:        version,
		BuildTimestamp: normalizedOrDefault(buildTimestamp, defaultBuildTimestamp),
		CommitHash:     normalizedOrDefault(commitHash, defaultCommitHash),
	}
}

func resolveVersion(tagName string) string {
	trimmedTag := strings.TrimSpace(tagName)
	if isUnsetTag(trimmedTag) {
		return verslib.GenerateFallbackVersion().Go
	}

	versions, err := verslib.CalculateFromString(trimmedTag)
	if err != nil || versions == nil || strings.TrimSpace(versions.Go) == "" {
		return verslib.GenerateFallbackVersion().Go
	}
	return versions.Go
}

func isUnsetTag(tagName string) bool {
	return tagName == "" || tagName == defaultTagName || tagName == defaultBuildTimestamp
}

func normalizedOrDefault(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
