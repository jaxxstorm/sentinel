## Why

Release and validation pipelines currently take too long in GitHub Actions when building across multiple OS/architecture targets, slowing feedback and releases. We need to reduce cross-platform CI build time without changing release artifact correctness.

## What Changes

- Optimize GitHub Actions cross-platform build execution to reduce wall-clock duration for release and validation workflows.
- Improve cache strategy for Go modules/build outputs and Docker Buildx layers in CI.
- Reduce redundant work between workflows and jobs (for example repeated setup, duplicate dry-runs, and unnecessary serial bottlenecks).
- Preserve current release outputs and metadata guarantees (binaries, checksums, GHCR images, version metadata).

## Capabilities

### New Capabilities
- `ci-cross-platform-build-optimization`: Define measurable performance requirements and optimization behavior for cross-platform CI builds.

### Modified Capabilities
- `goreleaser-binary-release`: Adjust workflow/build requirements to support faster multi-platform release execution while preserving binary artifacts.
- `ghcr-docker-publish-workflow`: Adjust image publish workflow requirements for faster multi-arch build/push behavior in GitHub Actions.

## Impact

- Affected systems: GitHub Actions workflows (`release.yml`, `release-validation.yml`, `publish-image.yml`) and potentially `.goreleaser.yaml`.
- Affected docs: release workflow/operator docs if behavior or constraints change.
- Risk areas: cache correctness, reproducibility, and release parity across platforms.
