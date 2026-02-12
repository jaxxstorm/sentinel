# tailscale-interactive-login-flow Specification

## Purpose
TBD - created by archiving change add-tailscale-node-auth-onboarding. Update Purpose after archive.
## Requirements
### Requirement: Sentinel SHALL provide an interactive enrollment path
Sentinel SHALL support interactive Tailscale login when configured login mode allows interactive enrollment.

#### Scenario: Interactive mode starts when no auth key is configured
- **WHEN** login mode is `interactive` and no auth key is provided
- **THEN** Sentinel initiates interactive enrollment and waits for completion

### Requirement: Sentinel SHALL present login instructions to operators
Sentinel SHALL display login URL or code instructions required by Tailscale enrollment in both pretty and JSON output modes.

#### Scenario: Login URL is presented in pretty output
- **WHEN** Sentinel enters interactive login state in pretty mode
- **THEN** output includes a human-readable login URL/code and clear next-step guidance

### Requirement: Sentinel SHALL enforce interactive login timeout and cancellation handling
Sentinel SHALL enforce configured login timeout and explicitly surface cancellation outcomes.

#### Scenario: Interactive enrollment times out
- **WHEN** login is not completed before configured timeout
- **THEN** Sentinel exits interactive flow with timeout status and actionable retry guidance

#### Scenario: Operator cancels interactive login
- **WHEN** interactive login is canceled by operator
- **THEN** Sentinel reports cancellation state and exits without starting polling

