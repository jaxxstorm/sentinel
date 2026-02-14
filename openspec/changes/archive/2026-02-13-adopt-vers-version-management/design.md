## Context

Sentinel currently surfaces version information through mutable package-level constants populated via GoReleaser `ldflags`. This works for tagged releases but is loosely defined for local builds and couples CLI output directly to constant names. The change introduces `github.com/jaxxstorm/vers` as the version metadata abstraction and standardizes how release metadata (tag, commit, build time) is injected and rendered by `sentinel version`.

## Goals / Non-Goals

**Goals:**
- Make `vers` the canonical version metadata source used by Sentinel runtime.
- Ensure release-tagged binaries report deterministic version metadata through `sentinel version`.
- Preserve useful fallback metadata for local/untagged builds.
- Keep command behavior backward-compatible (`sentinel version` remains simple and config-independent).

**Non-Goals:**
- Reworking unrelated CLI command output formatting.
- Introducing a new runtime config surface for version behavior.
- Changing release artifact topology (binary/image workflows stay as-is).

## Decisions

### Decision: Introduce an internal version adapter backed by vers
Sentinel will add an internal package (for example `internal/version`) that wraps `vers` and returns a typed struct used by CLI output.

Rationale:
- Avoid scattering `vers` calls across commands.
- Keep future output changes localized.
- Support testing without depending on direct CLI `stdout` parsing only.

Alternatives considered:
- Call `vers` directly in `internal/cli/version.go`: rejected due to tight command coupling.
- Keep constants-only model: rejected because it does not formalize metadata semantics.

### Decision: Keep build-time metadata injection in build tooling, but align names with vers usage
Release builds will continue setting metadata at build time via GoReleaser/CI, but targets and names will be aligned to the new version adapter so `sentinel version` reads a single coherent source.

Rationale:
- Retains existing release mechanics.
- Minimizes workflow churn while improving correctness.

Alternatives considered:
- Runtime `git` introspection: rejected for non-git/container runtime environments.

### Decision: Define deterministic fallback values for untagged builds
Local or snapshot builds must produce non-empty version output with documented fallback values.

Rationale:
- Improves debuggability and supportability.
- Prevents ambiguous empty output fields.

Alternatives considered:
- Fail or warn on missing metadata: rejected for local developer ergonomics.

## Risks / Trade-offs

- [Risk] `vers` API assumptions may drift from expected runtime shape. -> Mitigation: encapsulate usage in adapter package and unit test expected fields.
- [Risk] GoReleaser/CI metadata mapping mismatch could produce incomplete release output. -> Mitigation: add tests for version adapter defaults and release-path integration checks in CI.
- [Risk] Output changes could break downstream parsing expectations. -> Mitigation: keep field labels stable unless explicitly documented as changed.

## Migration Plan

1. Add `vers` dependency and create internal version adapter package.
2. Update `sentinel version` command to use adapter output.
3. Update build metadata wiring in `.goreleaser.yaml` and workflows as needed to populate adapter inputs.
4. Add/adjust tests for tagged and untagged metadata behavior.
5. Update docs if output semantics change.

Rollback strategy:
- Revert adapter wiring to previous constants model and restore prior `ldflags` mapping if issues are found in release validation.

## Open Questions

- Should `sentinel version` optionally support JSON output in this change, or stay human-only for now?
- Should the fallback version string be `dev`, `main`, or a `vers` default to reduce custom logic?
