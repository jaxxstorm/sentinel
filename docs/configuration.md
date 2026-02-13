# Configuration Reference

Sentinel loads YAML or JSON configuration and applies `SENTINEL_` environment overrides.

## File and Override Behavior

- File can be passed with `--config`.
- If no file is passed, Sentinel attempts `sentinel.yaml` then `sentinel.json`.
- Environment overrides use `SENTINEL_` prefix and map nested keys with underscores.
  - Example: `SENTINEL_POLL_INTERVAL=5s`

Precedence is deterministic:
1. defaults
2. config file values
3. scalar `SENTINEL_` env overrides
4. structured env overrides (when present):
   - `SENTINEL_DETECTORS`
   - `SENTINEL_DETECTOR_ORDER`
   - `SENTINEL_NOTIFIER_SINKS`
   - `SENTINEL_NOTIFIER_ROUTES`

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
  - supported sink `type` values: `stdout`, `debug`, `webhook`, `discord`
  - `discord` sinks require a non-empty webhook URL
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

## Environment Variable Matrix

### Common scalar overrides

| Env var | Config key |
| --- | --- |
| `SENTINEL_POLL_INTERVAL` | `poll_interval` |
| `SENTINEL_POLL_JITTER` | `poll_jitter` |
| `SENTINEL_POLL_BACKOFF_MIN` | `poll_backoff_min` |
| `SENTINEL_POLL_BACKOFF_MAX` | `poll_backoff_max` |
| `SENTINEL_SOURCE_MODE` | `source.mode` |
| `SENTINEL_POLICY_DEBOUNCE_WINDOW` | `policy.debounce_window` |
| `SENTINEL_POLICY_SUPPRESSION_WINDOW` | `policy.suppression_window` |
| `SENTINEL_POLICY_RATE_LIMIT_PER_MIN` | `policy.rate_limit_per_min` |
| `SENTINEL_POLICY_BATCH_SIZE` | `policy.batch_size` |
| `SENTINEL_OUTPUT_LOG_FORMAT` | `output.log_format` |
| `SENTINEL_OUTPUT_LOG_LEVEL` | `output.log_level` |
| `SENTINEL_OUTPUT_NO_COLOR` | `output.no_color` |
| `SENTINEL_TSNET_HOSTNAME` | `tsnet.hostname` |
| `SENTINEL_TSNET_STATE_DIR` | `tsnet.state_dir` |
| `SENTINEL_TSNET_LOGIN_MODE` | `tsnet.login_mode` |
| `SENTINEL_TSNET_AUTH_KEY` | `tsnet.auth_key` |
| `SENTINEL_TAILSCALE_AUTH_KEY` | Tailscale onboarding auth key fallback used by runtime wiring |
| `SENTINEL_TSNET_ALLOW_INTERACTIVE_FALLBACK` | `tsnet.allow_interactive_fallback` |
| `SENTINEL_TSNET_LOGIN_TIMEOUT` | `tsnet.login_timeout` |
| `SENTINEL_STATE_PATH` | `state.path` |
| `SENTINEL_CONFIG_PATH` | config file location used when `--config` is not set |

### Structured overrides (JSON values)

| Env var | Expected JSON shape | Overrides |
| --- | --- | --- |
| `SENTINEL_DETECTORS` | object map (`{"presence":{"enabled":true}}`) | full `detectors` map |
| `SENTINEL_DETECTOR_ORDER` | array (`["presence","runtime"]`) | full `detector_order` list |
| `SENTINEL_NOTIFIER_SINKS` | array of sink objects | full `notifier.sinks` list |
| `SENTINEL_NOTIFIER_ROUTES` | array of route objects | full `notifier.routes` list |

If a structured env key is malformed or empty, Sentinel fails startup with an error that includes the env key name.

### Tailscale auth key precedence

Sentinel resolves the onboarding auth key in this order:

1. `--tailscale-auth-key` CLI flag
2. `SENTINEL_TAILSCALE_AUTH_KEY`
3. `SENTINEL_TSNET_AUTH_KEY` / `tsnet.auth_key`

## Sink URL Environment Interpolation

Sink URLs support `${VAR_NAME}` expansion in config. Example:

```yaml
notifier:
  sinks:
    - name: webhook-primary
      type: webhook
      url: ${REQUESTBIN_WEBHOOK_URL}
    - name: discord-primary
      type: discord
      url: ${SENTINEL_DISCORD_WEBHOOK_URL}
```

Set the variable before running Sentinel:

```bash
export REQUESTBIN_WEBHOOK_URL="https://your-endpoint"
```

If the variable is missing, the URL resolves to empty and Sentinel skips that webhook sink.

Structured env values can also include `${VAR_NAME}` placeholders in sink URLs. Sentinel applies placeholder expansion after env parsing.

## Environment-Only Example

Sentinel can run without a config file if all required values are provided via environment:

```bash
SENTINEL_STATE_PATH=.sentinel/state.json \
SENTINEL_TSNET_STATE_DIR=.sentinel/tsnet \
SENTINEL_NOTIFIER_SINKS='[{"name":"stdout-debug","type":"stdout"},{"name":"discord-primary","type":"discord","url":"${SENTINEL_DISCORD_WEBHOOK_URL}"}]' \
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"sinks":["stdout-debug","discord-primary"]}]' \
go run ./cmd/sentinel validate-config
```

## Full Example

See [`config.example.yaml`](../config.example.yaml).
