# ghcr-docker-publish-workflow Specification

## Purpose
TBD - created by archiving change add-release-artifacts-goreleaser-ghcr. Update Purpose after archive.

## Requirements
### Requirement: Sentinel SHALL publish Docker images to GHCR from GitHub Actions
Sentinel SHALL provide a dedicated GitHub Actions workflow that builds the project Docker image and pushes it to GitHub Container Registry.

#### Scenario: Tagged release triggers GHCR image publish
- **WHEN** a semantic version tag is pushed
- **THEN** GitHub Actions builds the Docker image and pushes versioned tags to `ghcr.io/<owner>/<repo>`

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
