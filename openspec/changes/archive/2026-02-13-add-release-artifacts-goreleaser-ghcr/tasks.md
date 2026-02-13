## 1. GoReleaser Binary Release Setup

- [x] 1.1 Add `.goreleaser.yaml` with binary/archive/checksum configuration for Sentinel release artifacts.
- [x] 1.2 Ensure GoReleaser config excludes Docker/image publishing stanzas.
- [x] 1.3 Add a release workflow that runs GoReleaser on semantic version tags and publishes GitHub Release assets.

## 2. GHCR Docker Publish Workflow

- [x] 2.1 Add a dedicated Docker workflow that builds and pushes images to `ghcr.io/<owner>/<repo>` on semantic version tags.
- [x] 2.2 Configure Docker metadata/tag generation (version tags and OCI labels) in the GHCR workflow.
- [x] 2.3 Configure least-privilege workflow permissions for GHCR publishing (`packages: write`, `contents: read`).

## 3. Workflow Separation and Validation

- [x] 3.1 Ensure the GoReleaser workflow is binary-only and does not invoke Docker image publish steps.
- [x] 3.2 Ensure the Docker workflow operates independently of GoReleaser and can complete image publishing on tag events.
- [x] 3.3 Add/update CI checks or dry-run guidance to validate release and container workflows before production tags.

## 4. Documentation Updates

- [x] 4.1 Update Docsify documentation to describe the binary release flow (GoReleaser + GitHub Releases).
- [x] 4.2 Update Docsify documentation to describe the Docker image flow (GHCR workflow, image location, tagging).
- [x] 4.3 Add troubleshooting guidance for common release workflow failures (permissions, auth, missing tags, publish failures).

## 5. End-to-End Release Verification

- [x] 5.1 Verify workflow YAML and goreleaser config syntax in CI/local checks.
- [x] 5.2 Validate that release artifacts and GHCR images can be traced back to the same tag/revision metadata.
