# Troubleshooting

## Event appears in stdout sink but not webhook

1. Confirm webhook sink URL resolved from environment.
2. Look for sink logs:
   - `webhook send succeeded`
   - `webhook send failed`
3. Verify endpoint availability independently with `curl`.
4. Keep process running long enough to observe retry completion.

## Event appears in stdout sink but not Discord

1. Confirm the Discord sink has `type: discord` and a non-empty webhook URL.
2. Run config validation and check for sink errors:
   - `go run ./cmd/sentinel validate-config --config ./config.example.yaml`
3. Look for sink logs:
   - `discord send succeeded`
   - `discord send failed`
4. Confirm the webhook is a Discord webhook URL and not revoked.
5. Keep process running long enough to observe retry completion.

## Repeated transitions and idempotency

Sentinel stores idempotency keys in the state file.

- State path defaults to `.sentinel/state.json`.
- Duplicate suppression applies when the same idempotency key is seen inside TTL.
- Idempotency keys are derived from event attributes including timestamp, so repeated real transitions at different times are delivered.

Inspect current state:

```bash
cat .sentinel/state.json
```

## Config validation failures

Run:

```bash
go run ./cmd/sentinel validate-config --config ./config.example.yaml
```

Common issues:

- invalid durations
- missing `state.path` or `tsnet.state_dir`
- unsupported `source.mode` or `tsnet.login_mode`
- unknown `notifier.routes[].event_types` values (use a documented event type or `*`)
- unsupported `notifier.sinks[].type` values
- `discord` sink configured without a webhook URL
- malformed structured env JSON values (`SENTINEL_NOTIFIER_SINKS`, `SENTINEL_NOTIFIER_ROUTES`, etc.)

## Structured environment parsing errors

When using env-only deployments, Sentinel expects JSON for structured keys:

- `SENTINEL_DETECTORS`
- `SENTINEL_DETECTOR_ORDER`
- `SENTINEL_NOTIFIER_SINKS`
- `SENTINEL_NOTIFIER_ROUTES`

If startup fails with `parse SENTINEL_...`, check:

1. key value is valid JSON (not YAML)
2. arrays/objects use proper quoting for your shell
3. key is not set to an empty string
4. referenced detector names in `SENTINEL_DETECTOR_ORDER` exist in `SENTINEL_DETECTORS`

Quick local sanity check:

```bash
SENTINEL_NOTIFIER_SINKS='[{\"name\":\"stdout-debug\",\"type\":\"stdout\"}]' \
SENTINEL_NOTIFIER_ROUTES='[{\"event_types\":[\"*\"],\"sinks\":[\"stdout-debug\"]}]' \
go run ./cmd/sentinel validate-config
```

## Env precedence checks

Sentinel resolves config in this order:

1. defaults
2. config file values
3. scalar `SENTINEL_` env overrides
4. structured env keys (full-section overrides)

If behavior is surprising, check whether a structured key is replacing an entire section.

## Wildcard routing checks

If you expect all events to be delivered, verify route config contains wildcard:

```yaml
notifier:
  routes:
    - event_types: ["*"]
      sinks: ["stdout-debug", "webhook-primary"]
```

If `event_types` omits `*`, only explicitly listed event types are delivered.

## Migrating from presence-only routes

Previous configs commonly used:

```yaml
event_types: ["peer.online", "peer.offline"]
```

To include expanded event families, either:

1. Switch to wildcard:
   - `event_types: ["*"]`
2. Keep explicit control and add selected types:
   - `peer.routes.changed`
   - `daemon.state.changed`
   - `prefs.run_ssh.changed`

## Local verification workflow

```bash
go run ./cmd/sentinel validate-config --config ./config.example.yaml
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
go run ./cmd/sentinel test-notify --config ./config.example.yaml
```

```bash
go run ./cmd/sentinel test-notify --config ./config.example.yaml --dry-run
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
go run ./cmd/sentinel run --config ./config.example.yaml --log-level debug
```

## Release workflow failures

If tagged releases are not publishing artifacts:

1. Confirm tag format is semantic version (`vX.Y.Z` or prerelease `vX.Y.Z-rc.1`).
2. Confirm `Release Binaries` workflow has `contents: write`.
3. Check `.goreleaser.yaml` validity:
   ```bash
   goreleaser check
   ```
4. Validate a local snapshot build:
   ```bash
   goreleaser release --snapshot --clean --skip=publish
   ```

If GHCR images are not publishing:

1. Confirm `Publish Container Image` workflow has `packages: write`.
2. Confirm image path resolves to `ghcr.io/<owner>/<repo>`.
3. Validate Docker build locally:
   ```bash
   docker build -t sentinel:dry-run \
     --build-arg TAG_NAME=v0.0.0-dryrun \
     --build-arg COMMIT_HASH=$(git rev-parse HEAD) \
     --build-arg BUILD_TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
     .
   ```
4. Confirm workflow used the same tag and commit for both release assets and image labels.
