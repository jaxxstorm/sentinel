## 1. Canonical shorthand env parsing

- [x] 1.1 Update notifier shorthand env constants/resolution in `internal/config/config.go` so canonical route keys are explicit (`SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES`, `SENTINEL_NOTIFIER_ROUTE_SINKS`) and legacy aliases are mapped secondarily.
- [x] 1.2 Implement deterministic canonical-first resolution when both canonical and legacy alias keys are set for the same shorthand route field.
- [x] 1.3 Keep shorthand append behavior unchanged after alias normalization (no change to route/sink append semantics or default fallback behavior).

## 2. Validation and regression coverage

- [x] 2.1 Add config tests covering canonical-only shorthand route keys for event types and sinks.
- [x] 2.2 Add config tests covering legacy-alias-only behavior (`SENTINEL_NOTIFIER_ROUTE_EVENT_TYPE`, `SENTINEL_NOTIFIER_SINK`) to preserve compatibility.
- [x] 2.3 Add config tests covering mixed canonical+alias input to verify canonical-first precedence and deterministic effective config.

## 3. Documentation and migration guidance

- [x] 3.1 Update `docs/configuration.md` env matrix and shorthand sections to use canonical names in primary examples.
- [x] 3.2 Add migration notes in docs that map deprecated aliases to canonical keys and clarify precedence.
- [x] 3.3 Update related operator docs/examples (`docs/docker-image.md`, `docs/docker-compose.md`, `docs/troubleshooting.md`, `.env.example`) to keep env var naming consistent with canonical shorthand guidance.
