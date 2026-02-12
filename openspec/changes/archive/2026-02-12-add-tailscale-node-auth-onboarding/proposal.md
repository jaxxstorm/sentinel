## Why

Sentinel currently assumes a usable Tailscale identity but does not provide a complete onboarding path to become a tailnet node. This blocks first-run usability and makes deployment brittle when an auth key is not pre-provisioned.

## What Changes

- Add explicit node onboarding support so Sentinel can join a tailnet using either a provided auth key or an interactive Tailscale login flow.
- Add configuration and CLI options for auth key input, interactive login enablement, node hostname, and tsnet state directory behavior.
- Add startup logic that detects uninitialized/expired auth state and triggers the appropriate enrollment flow instead of failing silently.
- Add operator-facing status/output that clearly indicates enrollment state (not joined, login required, joining, joined, auth failed).
- Add secure handling of credentials and onboarding artifacts, including redaction and guidance to avoid leaking auth material.
- Add tests for both enrollment modes and failure paths (invalid key, key expiry, canceled interactive login, network errors).

## Capabilities

### New Capabilities

- `tailscale-node-onboarding`: Sentinel can join a tailnet as a tsnet-backed node using auth key or interactive browser/device login.
- `tailscale-auth-key-management`: Sentinel accepts auth keys from config/env/flags with validation and safe redaction behavior.
- `tailscale-interactive-login-flow`: Sentinel can present and manage an operator login URL/code flow when no auth key is available.
- `tailscale-enrollment-status-reporting`: Sentinel exposes clear CLI/log status for node enrollment lifecycle and actionable errors.

### Modified Capabilities

- None.

## Impact

- Affected code: startup/runtime wiring for tsnet initialization, CLI/config loading, logging/output messaging, and source bootstrap sequence.
- APIs/contracts: configuration schema and CLI flag set for onboarding and auth settings.
- Dependencies/systems: deeper integration with tsnet login/auth APIs and local state persistence behavior.
- Security/operations: handling of sensitive auth material, enrollment-time troubleshooting, and safer first-run experience for operators.
