## 1. Baseline and performance targets

- [x] 1.1 Capture current wall-clock baseline for release, release-validation, and publish-image workflows from recent GitHub Actions runs.
- [x] 1.2 Define explicit target thresholds for cross-platform build duration (cold and warm cache) and record them in docs.
- [x] 1.3 Add workflow summary output that reports relevant job durations for regression tracking.

## 2. Release validation workflow optimization

- [x] 2.1 Refactor `.github/workflows/release-validation.yml` into parallel jobs for independent checks (lint/config, GoReleaser snapshot, Docker dry-run).
- [x] 2.2 Remove unnecessary setup steps per job (for example QEMU where only native-platform build is executed).
- [x] 2.3 Ensure each validation job uses appropriate cache-aware setup (Go modules/build cache and/or Buildx cache).

## 3. Release and publish workflow tuning

- [x] 3.1 Audit and optimize `.github/workflows/release.yml` for cache-aware GoReleaser execution without changing release artifact outputs.
- [x] 3.2 Audit and optimize `.github/workflows/publish-image.yml` multi-platform build settings and cache scopes for faster repeat runs.
- [x] 3.3 Verify release metadata and image labels/tags remain unchanged after performance optimizations.

## 4. Docs and verification

- [x] 4.1 Update release documentation with CI performance strategy, baseline/target expectations, and troubleshooting guidance for cache misses.
- [x] 4.2 Run workflow/lint validation checks (for example actionlint and Goreleaser config checks) after refactor.
- [x] 4.3 Compare post-change workflow durations to baseline and confirm target improvements are met or document follow-up gaps.
