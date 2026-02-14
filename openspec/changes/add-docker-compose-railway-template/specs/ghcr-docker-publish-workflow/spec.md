## MODIFIED Requirements

### Requirement: Sentinel SHALL publish Docker images to GHCR from GitHub Actions
Sentinel SHALL provide a dedicated GitHub Actions workflow that builds the project Docker image and pushes it to GitHub Container Registry, SHALL optimize multi-platform build execution to reduce total publish duration while preserving tag and label outputs, and SHALL maintain a `latest` tag intended for compose template defaults in addition to versioned tags.

#### Scenario: Tagged release triggers GHCR image publish
- **WHEN** a semantic version tag is pushed
- **THEN** GitHub Actions builds the Docker image and pushes versioned tags to `ghcr.io/<owner>/<repo>`

#### Scenario: Default branch publish updates latest tag
- **WHEN** the configured publish workflow runs for the default branch path used for compose template defaults
- **THEN** the resulting GHCR image includes/updates the `latest` tag

#### Scenario: Multi-platform image build uses cache strategy
- **WHEN** GHCR publish workflow executes multi-platform buildx builds
- **THEN** Docker build cache configuration is enabled to reduce repeated build cost across runs
