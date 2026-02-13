# Release Artifacts

Sentinel publishes binaries and container images from the same semantic version tags, but through separate workflows.

## Binary Releases (GoReleaser + GitHub Releases)

- Workflow: `.github/workflows/release.yml`
- Trigger: push tag matching `v*.*.*` (and prerelease `v*.*.*-*`)
- Tooling: GoReleaser using `.goreleaser.yaml`
- Output:
  - platform archives (`linux`, `darwin`, `windows`)
  - `checksums.txt`
  - assets attached to the matching GitHub Release

GoReleaser is intentionally binary-only in this project. It does not build or publish Docker images.

## Container Releases (GHCR Workflow)

- Workflow: `.github/workflows/publish-image.yml`
- Trigger: push tag matching `v*.*.*` (and prerelease `v*.*.*-*`)
- Registry: `ghcr.io/<owner>/<repo>`
- Output tags:
  - semantic version tags (for example `v1.2.3`, `v1.2`, `v1`)
  - commit tag (`sha-<fullsha>`)
  - optional `latest` (for stable semver tags via Docker metadata action)

The image workflow runs independently from GoReleaser.

Runtime configuration for container users, including environment variables, is documented in [`Docker Image`](docker-image.md).

## Shared Version Metadata

Both pipelines use the same tag/ref and commit SHA:

- binaries embed:
  - `TagName`
  - `BuildTimestamp`
  - `CommitHash`
- container images include OCI labels:
  - `org.opencontainers.image.source`
  - `org.opencontainers.image.revision`
  - `org.opencontainers.image.version`

This gives a direct mapping from release assets and images back to the same source revision.

## Pre-Release Validation

Workflow: `.github/workflows/release-validation.yml`

Runs on pull requests and can be started manually. It validates:

1. workflow syntax (`actionlint`)
2. `.goreleaser.yaml` (`goreleaser check`)
3. GoReleaser snapshot build (`release --snapshot --skip=publish`)
4. Docker build dry-run (`docker/build-push-action` with `push: false`)

## Release Procedure

1. Merge release-ready changes to `main`.
2. Create and push a semver tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
3. Watch both workflows complete:
   - `Release Binaries`
   - `Publish Container Image`
4. Verify outputs:
   - GitHub Release assets present for `v0.1.0`
   - GHCR image tags published for `ghcr.io/<owner>/<repo>:v0.1.0`

## Common Failure Cases

- Missing `contents: write` permissions in release workflow:
  - GoReleaser cannot create/update GitHub Releases.
- Missing `packages: write` permissions in image workflow:
  - push to GHCR fails.
- Non-semver tags:
  - workflows are not triggered.
- Invalid `.goreleaser.yaml`:
  - `release-validation` fails at `goreleaser check`.
