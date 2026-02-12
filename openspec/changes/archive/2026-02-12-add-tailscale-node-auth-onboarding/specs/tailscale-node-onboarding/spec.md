## ADDED Requirements

### Requirement: Sentinel SHALL perform enrollment before netmap polling
Sentinel SHALL execute a Tailscale enrollment phase before starting the poll/diff notification pipeline.

#### Scenario: Enrollment gate blocks poll loop until joined
- **WHEN** Sentinel starts and the node is not yet authenticated
- **THEN** Sentinel does not begin netmap polling until enrollment reaches `joined` state

### Requirement: Sentinel SHALL support deterministic onboarding mode selection
Sentinel SHALL select onboarding mode in this order: existing authenticated state, auth key mode, interactive login mode, then fail with actionable configuration error.

#### Scenario: Existing authenticated state is reused
- **WHEN** Sentinel starts with a valid existing tsnet state directory
- **THEN** Sentinel skips key/login enrollment and proceeds as a joined node

#### Scenario: No usable mode produces explicit startup error
- **WHEN** Sentinel has no valid existing state, no auth key, and interactive mode disabled
- **THEN** Sentinel exits with a clear error explaining how to configure onboarding

### Requirement: Sentinel SHALL classify enrollment failures
Sentinel SHALL classify onboarding failures as retryable or non-retryable and expose the class to command exit behavior and status reporting.

#### Scenario: Invalid auth key is non-retryable
- **WHEN** enrollment fails due to an invalid or expired auth key
- **THEN** Sentinel marks the failure as non-retryable and reports operator action required

