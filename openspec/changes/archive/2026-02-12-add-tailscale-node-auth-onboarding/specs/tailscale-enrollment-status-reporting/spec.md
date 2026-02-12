## ADDED Requirements

### Requirement: Sentinel SHALL expose enrollment lifecycle status
Sentinel SHALL expose onboarding lifecycle states including `not_joined`, `login_required`, `joining`, `joined`, and `auth_failed`.

#### Scenario: Status command reports joined state
- **WHEN** Sentinel is authenticated and connected as a tailnet node
- **THEN** `sentinel status` reports `joined` with node identity metadata

### Requirement: Sentinel SHALL include enrollment diagnostics in status output
Sentinel SHALL include mode used, last enrollment error code, and recommended next action when state is not `joined`.

#### Scenario: Auth failure includes remediation hint
- **WHEN** enrollment is in `auth_failed` state
- **THEN** status output includes the error category and a concrete remediation hint

### Requirement: Sentinel SHALL support structured enrollment status logs
Sentinel SHALL emit structured enrollment status records in JSON log mode with stable field names.

#### Scenario: JSON mode includes enrollment fields
- **WHEN** Sentinel logs onboarding transitions in JSON mode
- **THEN** records include stable keys for status, mode, and error classification

### Requirement: Sentinel SHALL ensure status output does not leak secrets
Sentinel SHALL redact sensitive onboarding values (including auth key material) in status and log output.

#### Scenario: Status output redacts sensitive value
- **WHEN** status output references auth source details
- **THEN** raw auth key values are not present in output

