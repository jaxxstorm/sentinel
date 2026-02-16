## Why

Current operator-facing docs and README primarily demonstrate `go run` workflows, which biases usage toward source-based development and makes production-friendly install paths less obvious. Sentinel now publishes release binaries and Docker images, so docs should make binary and container workflows the default for operators.

## What Changes

- Rework README quick-start content to emphasize three execution paths: GitHub release binary, Docker image, and source-based development.
- Add explicit installation guidance for downloading Sentinel binaries from GitHub Releases, including OS/arch selection and executable setup.
- Update operator docs examples to prefer installed binary (`sentinel ...`) and Docker usage for runtime workflows, while keeping `go run` examples in clearly scoped development sections.
- Improve cross-linking between installation, release artifacts, and Docker deployment documentation so users can choose an execution path quickly.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `docsify-operator-guides`: Expand operator guidance requirements to include first-class installation and runtime pathways for GitHub Releases and Docker, plus README positioning updates away from `go run`-first flows.

## Impact

- Affected docs: `README.md`, `docs/getting-started.md`, `docs/commands.md`, and release/deployment docs under `docs/`.
- No runtime behavior, API, or dependency changes.
- May require updating command snippets and examples in multiple documentation pages for consistency.
