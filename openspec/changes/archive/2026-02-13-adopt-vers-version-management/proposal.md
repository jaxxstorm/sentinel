## Why

Sentinel currently reports version data from ad-hoc constants populated by build tooling, which makes version metadata behavior harder to standardize across local, CI, and release builds. Adopting `github.com/jaxxstorm/vers` and formalizing build-tag/version wiring will make the `sentinel version` output consistent and easier to maintain.

## What Changes

- Introduce `vers` as the source of truth for Sentinel build/version metadata used by the `version` command.
- Update version/build metadata injection so release tags and commit/build details are consistently available at runtime.
- Define deterministic fallback behavior for local/dev builds where release metadata is unavailable.
- Align release/build configuration and documentation with the new version metadata model.

## Capabilities

### New Capabilities
- `version-metadata-management`: Define how Sentinel derives and exposes version/build metadata using `vers` across dev and release builds.

### Modified Capabilities
- `sentinel-cli-config`: Extend CLI requirements so `sentinel version` reports standardized version/build metadata.
- `goreleaser-binary-release`: Update release build requirements to ensure tagged builds propagate version metadata expected by Sentinel runtime.

## Impact

- Affected code: CLI version command, version/constants package(s), and build wiring.
- Affected build/release config: `.goreleaser.yaml` and related CI/release workflows as needed.
- Dependency changes: add `github.com/jaxxstorm/vers`.
- Operator impact: `sentinel version` output format/content may change to a more structured and consistent model.
