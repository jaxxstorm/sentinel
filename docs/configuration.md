# Configuration Reference

Sentinel loads YAML or JSON configuration and applies `SENTINEL_` environment overrides.

## File and Override Behavior

- File can be passed with `--config`.
- If no file is passed, Sentinel attempts `sentinel.yaml` then `sentinel.json`.
- Environment overrides use `SENTINEL_` prefix and map nested keys with underscores.
  - Example: `SENTINEL_POLL_INTERVAL=5s`

## Top-Level Keys

### `poll_interval`
Base poll interval in polling mode. Duration string.

### `poll_jitter`
Randomized delay added to poll interval in polling mode.

### `poll_backoff_min` / `poll_backoff_max`
Backoff window used when source polling/watch operations fail.

### `source`
- `mode`: `realtime` (default) or `poll`

### `detectors`
Detector enablement map. Built-in detectors:
- `presence.enabled`
- `peer_changes.enabled`
- `runtime.enabled`

### `detector_order`
Ordered list of enabled detector names.

### `policy`
- `debounce_window`
- `suppression_window`
- `rate_limit_per_min`
- `batch_size`

### `notifier`
- `idempotency_key_ttl`
- `sinks`: sink definitions
- `routes`: routing rules by event type and severity
  - `event_types` supports explicit values (for example `peer.online`) and wildcard `*` (match all event types)

### `state`
- `path`: state file path
- `idempotency_key_ttl`: retention for stored idempotency keys

### `output`
- `log_format`: `pretty` or `json`
- `log_level`
- `no_color`

### `tsnet`
- `hostname`
- `state_dir`
- `login_mode`: `auto`, `auth_key`, `interactive`
- `auth_key`
- `allow_interactive_fallback`
- `login_timeout`

## Sink URL Environment Interpolation

Sink URLs support `${VAR_NAME}` expansion in config. Example:

```yaml
notifier:
  sinks:
    - name: webhook-primary
      type: webhook
      url: ${REQUESTBIN_WEBHOOK_URL}
```

Set the variable before running Sentinel:

```bash
export REQUESTBIN_WEBHOOK_URL="https://your-endpoint"
```

If the variable is missing, the URL resolves to empty and Sentinel skips that webhook sink.

## Full Example

See [`config.example.yaml`](../config.example.yaml).
