## MODIFIED Requirements

### Requirement: Sentinel SHALL produce release binaries via GoReleaser in GitHub Actions
Sentinel SHALL provide GoReleaser configuration and a GitHub Actions workflow that builds release binaries for supported platforms from tagged releases, and SHALL inject version metadata required by Sentinel runtime version reporting.

#### Scenario: Tagged release triggers GoReleaser workflow
- **WHEN** a semantic version tag is pushed
- **THEN** GitHub Actions runs GoReleaser to build and publish release binaries and checksums

#### Scenario: Tagged release binary includes runtime version metadata
- **WHEN** GoReleaser builds Sentinel for a semantic version tag
- **THEN** the produced binary embeds tag/version, commit, and build timestamp metadata expected by `sentinel version`
