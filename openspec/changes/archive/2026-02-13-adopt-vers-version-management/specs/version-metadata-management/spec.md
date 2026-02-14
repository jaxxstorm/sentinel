## ADDED Requirements

### Requirement: Sentinel SHALL source runtime version metadata from vers
Sentinel SHALL use `github.com/jaxxstorm/vers` as the canonical source for runtime version metadata exposed by Sentinel binaries.

#### Scenario: Version metadata is sourced from vers-backed values
- **WHEN** Sentinel is built with release metadata inputs
- **THEN** runtime version fields used by Sentinel are resolved through `vers` instead of ad-hoc constant assignment logic

### Requirement: Sentinel SHALL support tag-aware release metadata
Sentinel MUST support version metadata derived from build-time tag information so released binaries can report the release version that produced them.

#### Scenario: Tagged build reports tagged version
- **WHEN** Sentinel is built from a semantic release tag
- **THEN** the runtime version metadata includes that semantic tag value

### Requirement: Sentinel SHALL provide deterministic fallback metadata for non-release builds
Sentinel MUST provide non-empty fallback version metadata for local or CI snapshot builds where release tag metadata is unavailable.

#### Scenario: Untagged build has fallback version metadata
- **WHEN** Sentinel is built locally without release tag metadata
- **THEN** runtime version metadata resolves to documented fallback values rather than empty strings
