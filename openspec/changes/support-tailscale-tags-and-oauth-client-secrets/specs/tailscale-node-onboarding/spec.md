## ADDED Requirements

### Requirement: Sentinel SHALL apply configured advertise tags during enrollment
Sentinel SHALL apply configured Tailscale advertise tags to tsnet server setup before onboarding begins.

#### Scenario: Advertise tags are applied for auth-based enrollment
- **WHEN** an operator configures one or more valid advertise tags and Sentinel performs auth-key or OAuth enrollment
- **THEN** Sentinel starts tsnet with those tags configured for enrollment

## MODIFIED Requirements

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
