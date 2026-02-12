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

## Route Matching

Routes match by:

- `event_types` (for example `peer.online`, `peer.offline`)
- optional `severities`
- list of target sink names

Example:

```yaml
routes:
  - event_types: ["peer.online", "peer.offline"]
    severities: []
    sinks: ["stdout-debug", "webhook-primary"]
```

If a configured route has no available sinks at runtime, Sentinel falls back to `stdout-debug`.

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
