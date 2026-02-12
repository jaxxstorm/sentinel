## Context

Sentinel already embeds `tsnet` but currently depends on a pre-existing authenticated node state, which leaves first-run and re-auth scenarios undefined. This design adds an explicit onboarding/auth subsystem so Sentinel can become a functional Tailscale node by either using a provided auth key or initiating an interactive login flow, with clear operator status and safe secret handling.

## Goals / Non-Goals

**Goals:**

- Enable deterministic node enrollment at startup using either auth key or interactive login.
- Provide a single onboarding state machine used by `run` and diagnostic commands.
- Define secure input and redaction behavior for auth material across flags, env, config, and logs.
- Surface actionable enrollment status (pending login URL, joined identity, auth failures, retry hints).
- Keep onboarding compatible with existing polling/diff pipeline without redesigning core diff logic.

**Non-Goals:**

- Building a web UI hosted by Sentinel for authentication management.
- Supporting every possible Tailscale auth mode in v1 (focus on reusable auth keys + interactive login).
- Implementing multi-node orchestration or centralized secret distribution.
- Changing event diff semantics beyond startup precondition handling.

## Decisions

### 1. Introduce an explicit EnrollmentManager ahead of poll loop

Sentinel startup will call an `EnrollmentManager` before netmap polling begins. The manager encapsulates:

- existing-state detection (already authenticated node),
- auth key enrollment path,
- interactive login enrollment path,
- terminal statuses (`joined`, `login_required`, `auth_failed`, `retryable_error`).

Why this over implicit auth attempts scattered through source polling:
- creates a single lifecycle contract and test seam,
- prevents repeated noisy retries in unrelated poll code.

### 2. Define enrollment mode precedence and fallback behavior

Enrollment mode selection order:

1. If persisted tsnet state indicates already joined: use existing identity.
2. Else if auth key is configured (flag/env/config): attempt key-based auth.
3. Else if interactive login is enabled: trigger interactive login flow.
4. Else fail fast with actionable configuration error.

Fallback rules:
- Invalid/expired auth key does not silently fall back to interactive mode unless explicitly enabled by config (`allow_interactive_fallback`).
- Interactive login cancellation returns a distinct non-retryable status for operator action.

### 3. Add structured onboarding configuration surface

Configuration keys added under `tailscale`:

- `auth_key` (string, optional),
- `login_mode` (`auth_key|interactive|auto`),
- `allow_interactive_fallback` (bool),
- `hostname` (string),
- `state_dir` (string),
- `login_timeout` (duration).

CLI/env bindings:

- `--tailscale-auth-key`, `--tailscale-login-mode`, `--tailscale-state-dir`,
- `SENTINEL_TAILSCALE_AUTH_KEY`, `SENTINEL_TAILSCALE_LOGIN_MODE`, etc.

Auth key input precedence: CLI > env > config file.

### 4. Add sensitive-data guardrails and redaction defaults

- Auth key values are never logged in plain text.
- Logs expose only masked fingerprints (for troubleshooting duplicates/rotation).
- `dump-netmap` and status output avoid leaking onboarding secrets.
- Validation errors refer to source (`env`, `flag`, `config`) but not raw secret content.

### 5. Implement interactive login UX as URL/code output, not embedded page

Sentinel will rely on Tailscale-provided login link/code and print a deterministic, operator-friendly enrollment message in pretty or JSON mode.

- Pretty mode uses highlighted sections (`Login URL`, `Expires`, `Next step`).
- JSON mode emits structured enrollment status records.
- Command flow blocks until success, timeout, cancellation, or unrecoverable error.

Alternative considered:
- Hosting a local login web page in Sentinel. Rejected to avoid new web server attack surface and session management complexity.

### 6. Integrate enrollment status into health/status commands

`sentinel status` will include:

- node identity if joined (node ID/hostname),
- enrollment mode used,
- current auth state and last error code if not joined.

`run --once` returns non-zero when enrollment cannot complete.

## Risks / Trade-offs

- [Interactive login may block unattended environments] -> Mitigation: require explicit interactive mode and clear failure if running non-interactively.
- [Auth key rotation/expiry can break restarts unexpectedly] -> Mitigation: explicit error classification and retry guidance; support reloading key via env/config.
- [Secret leakage via debug output] -> Mitigation: centralized redaction helpers and tests asserting masked output.
- [Enrollment retries may delay startup] -> Mitigation: bounded retry/backoff with configurable timeout and immediate fail-fast modes.
- [Platform differences in opening login URLs] -> Mitigation: print URL/code always; never depend on auto-open browser success.

## Migration Plan

1. Add onboarding config schema, validation, and CLI/env bindings.
2. Implement `EnrollmentManager` with mode selection and status model.
3. Integrate manager into runtime startup path before source polling.
4. Add pretty/JSON enrollment status output and status command extensions.
5. Add tests for key auth success/failure, interactive success/timeout/cancel, and redaction behavior.
6. Rollout: default to `auto` mode with safe fail-fast messaging for unconfigured auth in non-interactive contexts.

Rollback:

- Disable onboarding gating with a temporary compatibility flag (if needed) and revert to previous release.
- Preserve state directory format compatibility to avoid destructive rollback steps.

## Open Questions

- Should interactive login be permitted by default in `auto` mode for local terminals, or require explicit opt-in everywhere?
- Do we need hot-reload for rotated auth keys without restarting Sentinel?
- What minimum status fields should be exposed for `status --json` to support automation?
