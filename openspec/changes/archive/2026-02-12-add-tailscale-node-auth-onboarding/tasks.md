## 1. Enrollment foundation and interfaces

- [x] 1.1 Add onboarding state model (`not_joined`, `login_required`, `joining`, `joined`, `auth_failed`) and error-classification types
- [x] 1.2 Create `EnrollmentManager` interface and default implementation entrypoint used by runtime startup
- [x] 1.3 Wire enrollment phase to execute before netmap polling starts

## 2. Configuration and CLI wiring

- [x] 2.1 Add `tailscale` config fields for `auth_key`, `login_mode`, `allow_interactive_fallback`, `state_dir`, `hostname`, and `login_timeout`
- [x] 2.2 Add CLI flags and env mappings for onboarding settings with documented precedence
- [x] 2.3 Implement config validation for mode combinations and required values

## 3. Auth key onboarding path

- [x] 3.1 Implement auth key source resolution precedence (flag > env > config)
- [x] 3.2 Implement key-based tsnet enrollment call path and success transition to `joined`
- [x] 3.3 Implement invalid/expired key classification as non-retryable `auth_failed`
- [x] 3.4 Implement optional interactive fallback gating after key failure

## 4. Interactive login onboarding path

- [x] 4.1 Implement interactive login flow trigger when mode permits and no usable auth key exists
- [x] 4.2 Implement timeout and cancellation handling for interactive enrollment
- [x] 4.3 Emit operator login instructions (URL/code/next-step) in pretty and JSON output modes
- [x] 4.4 Ensure interactive flow exits without starting poll loop on unresolved enrollment

## 5. Startup mode selection and existing-state reuse

- [x] 5.1 Implement deterministic mode selection order (existing state -> auth key -> interactive -> fail)
- [x] 5.2 Detect valid existing tsnet authenticated state and bypass new enrollment
- [x] 5.3 Implement explicit startup failure when no permitted onboarding mode is available

## 6. Status and logging integration

- [x] 6.1 Extend `sentinel status` output with enrollment state, mode used, and node identity when joined
- [x] 6.2 Add structured JSON log fields for onboarding transitions and error categories
- [x] 6.3 Add remediation hints in status output for non-joined states

## 7. Security and redaction controls

- [x] 7.1 Add centralized auth-key redaction helper for logs, errors, and status output
- [x] 7.2 Ensure validation and onboarding errors never include raw auth key values
- [x] 7.3 Add tests that assert no auth material appears in emitted output

## 8. Runtime behavior and command semantics

- [x] 8.1 Ensure `run --once` returns non-zero when enrollment cannot complete
- [x] 8.2 Keep polling/diff pipeline blocked until `joined` state is reached
- [x] 8.3 Add enrollment diagnostics to startup logs without leaking secrets

## 9. Test coverage

- [x] 9.1 Add unit tests for onboarding mode selection and precedence logic
- [x] 9.2 Add tests for auth key success and invalid/expired key failure classification
- [x] 9.3 Add tests for interactive flow success, timeout, and cancellation behavior
- [x] 9.4 Add tests for status output fields and remediation hints
- [x] 9.5 Add integration test covering first-run onboarding to joined state before first poll
