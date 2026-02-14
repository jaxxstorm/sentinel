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

Sentinel runtime version data is normalized through `github.com/jaxxstorm/vers`.

Both pipelines use the same tag/ref and commit SHA, then inject build metadata consumed by `sentinel version`:

- binaries inject:
  - `internal/version.TagName`
  - `internal/version.BuildTimestamp`
  - `internal/version.CommitHash`
- container images include OCI labels:
  - `org.opencontainers.image.source`
  - `org.opencontainers.image.revision`
  - `org.opencontainers.image.version`

This gives a direct mapping from release assets and images back to the same source revision.

In non-release local builds (for example `go run` without tag ldflags), Sentinel falls back to vers defaults such as `v0.0.0-dev`.

## Pre-Release Validation

Workflow: `.github/workflows/release-validation.yml`

Runs on pull requests and can be started manually. It validates in parallel jobs:

1. workflow syntax (`actionlint`)
2. `.goreleaser.yaml` (`goreleaser check`) and GoReleaser snapshot build (`release --snapshot --skip=publish`)
3. Docker build dry-run (`docker/build-push-action` with `push: false`)

## CI Performance Baseline and Targets

Baseline values below were captured from recent GitHub Actions runs before workflow optimization (February 2026).
Targets are enforced as warnings in workflow summaries and used for regression tracking.

### Workflow-level targets

| Workflow | Baseline cold | Baseline warm | Target cold | Target warm |
| --- | ---: | ---: | ---: | ---: |
| `release-validation` | 900s | 760s | 420s | 300s |
| `release` | 720s | 610s | 600s | 450s |
| `publish-image` | 960s | 690s | 540s | 360s |

### Job-level targets

| Job | Target cold | Target warm | Notes |
| --- | ---: | ---: | --- |
| `workflow-lint` | 90s | 90s | No Go or container setup. |
| `goreleaser-validate` | 420s | 300s | Uses Go dependency/build cache. |
| `docker-dry-run` | 240s | 150s | Native `linux/amd64` only, no QEMU. |

### Post-change comparison

- `release-validation` critical path target (`<=420s`) is a 53% reduction from the 900s cold baseline.
- `release` target (`<=600s`) is a 17% reduction from the 720s cold baseline.
- `publish-image` target (`<=540s`) is a 44% reduction from the 960s cold baseline.

If a run exceeds target, the workflows emit a warning in the job summary and should be treated as a performance regression to investigate.

## Cache Strategy and Troubleshooting

Sentinel uses cache-aware setup in all release workflows:

- Go module/build cache via `actions/setup-go` and `actions/cache`.
- Docker Buildx layer cache via `cache-from/cache-to` with GHA scopes.
- Cross-workflow reuse by reading `release-validation-dry-run` cache in `publish-image`.

If build times regress:

1. Check `GITHUB_STEP_SUMMARY` in the workflow run for elapsed duration and target.
2. Confirm `go.sum` or `Dockerfile` changes did not invalidate caches unexpectedly.
3. Verify Buildx cache scopes (`publish-image`, `release-validation-dry-run`) are unchanged.
4. Re-run once to separate cold-cache behavior from persistent regressions.
5. If warm runs still exceed target, open a follow-up change and capture updated baseline numbers.

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
