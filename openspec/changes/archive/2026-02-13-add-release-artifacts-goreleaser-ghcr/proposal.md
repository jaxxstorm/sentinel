## Why

Sentinel needs a repeatable, automated release process for binaries and container images so tagged releases are consistent and easy to consume. Today there is no standardized release automation using GoReleaser and GitHub Actions.

## What Changes

- Add release automation for Sentinel binaries using GoReleaser executed by GitHub Actions.
- Add a separate GitHub Actions workflow that builds and publishes a Docker image to GitHub Container Registry (GHCR).
- Ensure Docker publishing is explicitly handled outside GoReleaser (GoReleaser builds binaries/artifacts only).
- Define release trigger, tagging, permissions, and artifact naming expectations for both workflows.
- Add operator/developer documentation describing how to cut a release and what outputs are produced.

## Capabilities

### New Capabilities
- `goreleaser-binary-release`: Build and publish versioned Sentinel binary release artifacts via GoReleaser in GitHub Actions.
- `ghcr-docker-publish-workflow`: Build and push Sentinel Docker images to GHCR in a dedicated GitHub Actions workflow, independent of GoReleaser.

### Modified Capabilities
- `docsify-operator-guides`: Add release process documentation covering binary and container artifact flows.

## Impact

- Affected code/config: `.goreleaser.yaml` (new), `.github/workflows/*` release/image workflows, `Dockerfile` usage in CI, and release docs.
- External systems: GitHub Releases and `ghcr.io/<owner>/<repo>` container publishing.
- Security/permissions: GitHub Actions permissions for release creation and package push.
- Developer workflow: tagged releases become the canonical mechanism for producing distributable binaries and container images.
