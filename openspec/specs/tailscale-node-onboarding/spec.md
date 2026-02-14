# tailscale-node-onboarding Specification

## Purpose
TBD - created by archiving change add-tailscale-node-auth-onboarding. Update Purpose after archive.
## Requirements
### Requirement: Sentinel SHALL perform enrollment before netmap polling
Sentinel SHALL execute a Tailscale enrollment phase before starting the poll/diff notification pipeline.

#### Scenario: Enrollment gate blocks poll loop until joined
- **WHEN** Sentinel starts and the node is not yet authenticated
- **THEN** Sentinel does not begin netmap polling until enrollment reaches `joined` state

### Requirement: Sentinel SHALL support deterministic onboarding mode selection
Sentinel SHALL select onboarding mode in this order: existing authenticated state, auth key mode, OAuth credential mode, interactive login mode, then fail with actionable configuration error.

#### Scenario: Existing authenticated state is reused
- **WHEN** Sentinel starts with a valid existing tsnet state directory
- **THEN** Sentinel skips key/login enrollment and proceeds as a joined node

#### Scenario: OAuth credential mode is selected when auth key is unavailable
- **WHEN** Sentinel has no reusable existing state, no auth key source, and valid OAuth credential inputs are configured
- **THEN** Sentinel attempts onboarding using OAuth credential mode before interactive login

#### Scenario: No usable mode produces explicit startup error
- **WHEN** Sentinel has no valid existing state, no auth key, no valid OAuth credentials, and interactive mode disabled
- **THEN** Sentinel exits with a clear error explaining how to configure onboarding

### Requirement: Sentinel SHALL apply configured advertise tags during enrollment
Sentinel SHALL apply configured Tailscale advertise tags to tsnet server setup before onboarding begins.

#### Scenario: Advertise tags are applied for auth-based enrollment
- **WHEN** an operator configures one or more valid advertise tags and Sentinel performs auth-key or OAuth enrollment
- **THEN** Sentinel starts tsnet with those tags configured for enrollment

### Requirement: Sentinel SHALL classify enrollment failures
Sentinel SHALL classify onboarding failures as retryable or non-retryable and expose the class to command exit behavior and status reporting.

#### Scenario: Invalid auth key is non-retryable
- **WHEN** enrollment fails due to an invalid or expired auth key
- **THEN** Sentinel marks the failure as non-retryable and reports operator action required
