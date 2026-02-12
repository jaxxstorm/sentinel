# tailscale-enrollment-status-reporting Specification

## Purpose
TBD - created by archiving change add-tailscale-node-auth-onboarding. Update Purpose after archive.

## Requirements
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
Sentinel SHALL emit structured enrollment status records in JSON log mode with stable field names, MUST include source attribution for enrollment records, and MUST omit `error_code` and `error_class` when those values are empty.

#### Scenario: JSON mode includes enrollment fields
- **WHEN** Sentinel logs onboarding transitions in JSON mode
- **THEN** records include stable keys for status and mode

#### Scenario: Empty enrollment error fields are omitted
- **WHEN** enrollment status has no error classification values
- **THEN** runtime logs do not include empty-string `error_code` or `error_class` fields

#### Scenario: Enrollment log includes source attribution
- **WHEN** Sentinel emits an onboarding status log record
- **THEN** the record includes a stable `log_source` value for origin filtering

### Requirement: Sentinel SHALL ensure status output does not leak secrets
Sentinel SHALL redact sensitive onboarding values (including auth key material) in status and log output.

#### Scenario: Status output redacts sensitive value
- **WHEN** status output references auth source details
- **THEN** raw auth key values are not present in output
