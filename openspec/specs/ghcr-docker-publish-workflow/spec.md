# ghcr-docker-publish-workflow Specification

## Purpose
TBD - created by archiving change add-release-artifacts-goreleaser-ghcr. Update Purpose after archive.

## Requirements
### Requirement: Sentinel SHALL publish Docker images to GHCR from GitHub Actions
Sentinel SHALL provide a dedicated GitHub Actions workflow that builds the project Docker image and pushes it to GitHub Container Registry, SHALL optimize multi-platform build execution to reduce total publish duration while preserving tag and label outputs, and SHALL maintain a `latest` tag intended for compose template defaults in addition to versioned tags.

#### Scenario: Tagged release triggers GHCR image publish
- **WHEN** a semantic version tag is pushed
- **THEN** GitHub Actions builds the Docker image and pushes versioned tags to `ghcr.io/<owner>/<repo>`

#### Scenario: Default branch publish updates latest tag
- **WHEN** the configured publish workflow runs for the default branch path used for compose template defaults
- **THEN** the resulting GHCR image includes/updates the `latest` tag

#### Scenario: Multi-platform image build uses cache strategy
- **WHEN** GHCR publish workflow executes multi-platform buildx builds
- **THEN** Docker build cache configuration is enabled to reduce repeated build cost across runs

### Requirement: GHCR workflow SHALL be independent of GoReleaser image building
Container build and publish actions SHALL be implemented in the Docker workflow and SHALL NOT depend on GoReleaser image features.

#### Scenario: Docker workflow builds image without goreleaser docker stanza
- **WHEN** the release pipeline runs
- **THEN** the GHCR workflow builds/pushes images directly and GoReleaser remains binary-only

### Requirement: GHCR publishing SHALL include traceable image metadata
Published images SHALL include OCI labels and tags that map back to release version and source revision.

#### Scenario: Published image contains release metadata
- **WHEN** a release image is pushed to GHCR
- **THEN** image metadata includes source repository and revision labels, and tags include the release version
