# goreleaser-binary-release Specification

## Purpose
TBD - created by archiving change add-release-artifacts-goreleaser-ghcr. Update Purpose after archive.

## Requirements
### Requirement: Sentinel SHALL produce release binaries via GoReleaser in GitHub Actions
Sentinel SHALL provide GoReleaser configuration and a GitHub Actions workflow that builds release binaries for supported platforms from tagged releases.

#### Scenario: Tagged release triggers GoReleaser workflow
- **WHEN** a semantic version tag is pushed
- **THEN** GitHub Actions runs GoReleaser to build and publish release binaries and checksums

### Requirement: GoReleaser binary release SHALL exclude container image publishing
The GoReleaser configuration SHALL NOT build or publish Docker/OCI images.

#### Scenario: GoReleaser run emits no image artifacts
- **WHEN** the release workflow executes GoReleaser
- **THEN** produced artifacts include binaries/archives/checksums and do not include container image publish steps

### Requirement: Release binaries SHALL be discoverable through GitHub Releases
Binary artifacts produced by GoReleaser SHALL be attached to the corresponding GitHub Release for each version tag.

#### Scenario: Release assets are attached to tag release
- **WHEN** GoReleaser completes successfully for a tag
- **THEN** the GitHub Release for that tag contains the generated binary artifacts and checksums
