## Context

The pain point is GitHub Actions cross-platform build latency, not local Docker startup. Current release validation runs all checks in one job and performs setup steps that are not always required for the specific build target. Release and image publish workflows already use some caching, but there is no explicit performance baseline or guardrail to detect regressions.

## Goals / Non-Goals

**Goals:**
- Reduce cross-platform CI wall-clock time for release and validation workflows.
- Keep release outputs unchanged (same binaries, checksums, tags, labels, version metadata).
- Improve cache hit rates and remove unnecessary setup work.
- Add lightweight observability for workflow duration trends.

**Non-Goals:**
- Redesigning release artifact formats.
- Replacing GoReleaser or Docker Buildx with new tooling.
- Optimizing local developer `docker build` flow in this change.

## Decisions

### Decision: Split release validation into parallel jobs
Instead of a single serial `validate` job, independent checks will run in separate jobs (for example workflow lint/config checks vs snapshot binary build vs docker dry-run).

Rationale:
- Reduces total wall-clock time via parallel execution.
- Isolates failures and makes slow stages explicit.

Alternatives considered:
- Keep one serial validation job: rejected due to unavoidable cumulative runtime.

### Decision: Tighten setup and caching per job
Each job will only execute setup steps needed for its target. Examples include avoiding QEMU setup for single native-platform dry-runs and ensuring Go/setup actions use dependency-aware caching consistently.

Rationale:
- Prevents repeated setup overhead.
- Improves warm-run performance without changing outputs.

Alternatives considered:
- Global setup across all jobs: not feasible because GitHub Actions jobs are isolated.

### Decision: Preserve release artifact semantics while optimizing execution strategy
Optimization changes must not alter release output contracts (binary contents, checksums, tag/label behavior).

Rationale:
- Performance changes should be low-risk and operationally safe.

Alternatives considered:
- Aggressive platform reduction: rejected because it would change release support guarantees.

### Decision: Add explicit timing visibility in workflow output
Workflows will include step/job summaries to make duration data visible for baseline tracking and regression detection.

Rationale:
- Performance work needs observable feedback loops.

Alternatives considered:
- No measurement output: rejected because regressions would be hard to detect.

## Risks / Trade-offs

- [Risk] Parallel jobs can increase total runner minutes. -> Mitigation: prioritize wall-clock reduction for developer feedback loops and monitor cost.
- [Risk] Cache misuse could produce stale outputs. -> Mitigation: use scoped cache keys tied to lockfiles and build inputs.
- [Risk] Workflow refactors might accidentally alter release behavior. -> Mitigation: validate artifact parity in release-validation and keep changes incremental.

## Migration Plan

1. Capture baseline timing from current release/release-validation runs.
2. Refactor `release-validation.yml` into parallel jobs with minimal setup per job.
3. Audit and tune cache usage in release/image workflows.
4. Add workflow summary timing output for key jobs.
5. Validate output parity and compare duration against baseline.

Rollback strategy:
- Revert workflow files to previous serial configuration if regressions in release correctness or major instability are observed.

## Open Questions

- Should duration thresholds be hard fail gates or informational warnings initially?
- Do we want to optimize for wall-clock time only, or also set a runner-minute budget target?
