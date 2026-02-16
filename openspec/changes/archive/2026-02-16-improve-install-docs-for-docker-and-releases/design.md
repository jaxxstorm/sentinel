## Context

Sentinel documentation currently includes multiple `go run` examples in README and operator-facing pages. That is useful for contributors, but it is not the lowest-friction path for operators who want to install and run Sentinel quickly. Sentinel now has two production-oriented distribution channels (GitHub Release binaries and GHCR Docker images), and docs should lead with those channels.

The change must remain documentation-only: no CLI, config, runtime, or workflow behavior should change.

## Goals / Non-Goals

**Goals:**
- Make installation and execution paths explicit for three audiences: operators using release binaries, operators using Docker, and developers running from source.
- Add concrete GitHub Releases download instructions (OS/arch-aware examples, extraction, executable setup).
- Reorder and rewrite README quick start so `go run` is not the default operational path.
- Keep documentation internally consistent by aligning command examples across README and docs pages.

**Non-Goals:**
- Changing Sentinel runtime behavior, flags, config schema, or packaging outputs.
- Introducing a package-manager installation path (Homebrew, apt, etc.).
- Rewriting all docs pages; only pages that currently shape install/run workflows are in scope.

## Decisions

### 1. Prioritize install-first command hierarchy in user-facing docs
- Decision: Operator docs will prefer `sentinel ...` commands (installed binary) and Docker invocations for runtime examples; `go run` examples move to development-focused sections.
- Rationale: This matches how most operators deploy services and avoids requiring Go toolchain setup for basic usage.
- Alternative considered: Keep mixed command styles throughout docs. Rejected because it perpetuates ambiguity about recommended usage.

### 2. Add an explicit GitHub Releases installation flow
- Decision: Document download and setup from GitHub Releases with platform-aware examples and verification guidance.
- Rationale: Release binaries are already part of artifact strategy and should be discoverable from README/getting-started.
- Alternative considered: Link only to release page without commands. Rejected because users still must infer file selection and install steps.

### 3. Keep Docker guidance as a first-class execution path
- Decision: README and getting-started docs will include Docker path references alongside release binaries, with links to existing compose/deployment docs.
- Rationale: Many users prefer containerized execution and already have documented compose templates.
- Alternative considered: Keep Docker only in dedicated docker docs. Rejected because quick-start readers should see Docker immediately.

### 4. Scope updates to a focused doc set
- Decision: Update `README.md`, getting-started/install docs, and command examples most visible to new users; avoid broad content churn.
- Rationale: Minimizes regression risk and keeps the change reviewable.
- Alternative considered: Global rewrite of all docs command snippets. Rejected for size and risk.

## Risks / Trade-offs

- [Risk] Release asset naming or examples drift from actual published artifacts.
  - Mitigation: Anchor instructions to the existing release workflow naming conventions and include a fallback “check release assets list” step.
- [Risk] Users may interpret reduced `go run` visibility as dropped source support.
  - Mitigation: Preserve a clear development section that explicitly documents source-based workflows.
- [Risk] Inconsistent command styles remain in untouched pages.
  - Mitigation: Include a task to sweep high-traffic docs pages for command consistency in this change scope.

## Migration Plan

1. Update spec requirements for operator docs/readme expectations.
2. Edit README quick-start/install sections to include release binary and Docker-first paths.
3. Add/refresh installation guidance in docs for GitHub Releases download and binary setup.
4. Update selected docs command examples from `go run` to installed binary where operator-facing.
5. Verify cross-links among README, getting-started, release docs, and docker docs.

Rollback: Revert documentation commits if guidance conflicts are found; no data or runtime migration is required.

## Open Questions

- Should checksum verification be mandatory in examples or shown as an optional hardening step?
- Should we add a short matrix indicating when to choose release binary vs Docker vs source run?
