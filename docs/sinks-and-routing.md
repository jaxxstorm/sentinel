# Sinks and Routing

Sentinel emits events through notifier routes to one or more sinks.

## Sink Types

### `stdout` / `debug`
Writes machine-readable JSON to stdout.

Output shape:

```json
{
  "log_source": "sink",
  "sink": "stdout-debug",
  "event": {"event_type": "peer.online"},
  "idempotency_key": "..."
}
```

### `webhook`
Sends HTTP POST requests with JSON payload and `Idempotency-Key` header.

- Retries on failure (bounded attempts with backoff)
- Logs success/failure with sink name and status code

Success log example:

```text
INFO webhook send succeeded {"log_source":"sink","sink":"webhook-primary","status_code":200}
```

Failure log example:

```text
WARN webhook send failed {"log_source":"sink","sink":"webhook-primary","status_code":502,"attempt":2,"max_attempts":4}
```

### `discord`
Sends HTTP POST requests to a Discord webhook endpoint.

- Uses a Discord-friendly `content` payload with event summary fields.
- Includes `Idempotency-Key` header.
- Retries on failure (bounded attempts with backoff).
- Logs success/failure with sink name and status code.

Success log example:

```text
INFO discord send succeeded {"log_source":"sink","sink":"discord-primary","status_code":204}
```

Failure log example:

```text
WARN discord send failed {"log_source":"sink","sink":"discord-primary","status_code":502,"attempt":1,"max_attempts":4}
```

## Route Matching

Routes match by:

- `event_types` (explicit values or `*` for all event types)
- optional `severities`
- list of target sink names

Example:

```yaml
routes:
  - event_types: ["*"]
    severities: []
    sinks: ["stdout-debug", "webhook-primary", "discord-primary"]
```

Explicit matching is still supported:

```yaml
routes:
  - event_types: ["peer.online", "peer.offline", "peer.routes.changed"]
    sinks: ["webhook-primary"]
```

If `*` appears in a route's `event_types`, that route is treated as match-all even when literal values are also present.

If a configured route has no available sinks at runtime, Sentinel falls back to `stdout-debug`.

## Event Type Catalog

Current event families include:

- `peer.online`, `peer.offline`, `peer.added`, `peer.removed`
- `peer.routes.changed`, `peer.tags.changed`
- `peer.machine_authorized.changed`, `peer.key_expiry.changed`, `peer.key_expired`, `peer.hostinfo.changed`
- `daemon.state.changed`
- `prefs.advertise_routes.changed`, `prefs.exit_node.changed`, `prefs.run_ssh.changed`, `prefs.shields_up.changed`
- `tailnet.domain.changed`, `tailnet.tka_enabled.changed`

## Dry-Run Validation

Use `test-notify --dry-run` to validate route matching without external delivery:

```bash
go run ./cmd/sentinel test-notify --config ./config.example.yaml --dry-run
```

Use normal `test-notify` to validate actual webhook delivery:

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
go run ./cmd/sentinel test-notify --config ./config.example.yaml
```

To validate Discord delivery, point a Discord sink URL at your webhook:

```bash
SENTINEL_DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..." \
go run ./cmd/sentinel run --config ./config.example.yaml --log-level debug
```
