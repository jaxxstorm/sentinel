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
  - optional `device` selector narrows peer/device-scoped events:
    - `device.names`: match by device name
    - `device.tags`: match when any configured tag is present
    - `device.owners`: match by stable owner identity values
    - `device.ips`: match by literal IP addresses
  - selector semantics are deterministic:
    - OR within each selector field (`names`, `tags`, `owners`, `ips`)
    - AND across configured selector fields
    - routes with `device` selectors do not match non-device events (for example `daemon.state.changed`)

### `state`
- `path`: state file path
- `idempotency_key_ttl`: retention for stored idempotency keys

### `output`
- `log_format`: `pretty` or `json`
- `log_level`
- `no_color`

### `tsnet`
- `runtime_mode`: `embedded` (default) or `localapi`
- `localapi_socket`: tailscaled LocalAPI socket path (used when `runtime_mode=localapi`)
- `hostname`
- `state_dir`
- `advertise_tags`: list of tags in `tag:<name>` format
- `login_mode`: `auto`, `auth_key`, `oauth`, `interactive`
- `auth_key`
- `client_secret`
- `client_id`
- `id_token`
- `audience`
- `allow_interactive_fallback`
- `login_timeout`

`client_secret` requires `client_id`. If OAuth fields are set without `client_secret`, config validation fails.
In `runtime_mode=localapi`, Sentinel reads from an existing tailscaled socket and does not require tsnet onboarding credentials.

## Environment Variable Matrix

For required/optional semantics as used by repository compose templates (including Railway import),
see [Docker Compose and Railway](docker-compose.md).

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
| `SENTINEL_TSNET_RUNTIME_MODE` | `tsnet.runtime_mode` |
| `SENTINEL_TSNET_LOCALAPI_SOCKET` | `tsnet.localapi_socket` |
| `SENTINEL_TSNET_STATE_DIR` | `tsnet.state_dir` |
| `SENTINEL_TSNET_ADVERTISE_TAGS` | `tsnet.advertise_tags` (JSON array or comma-separated list) |
| `SENTINEL_TSNET_LOGIN_MODE` | `tsnet.login_mode` |
| `SENTINEL_TSNET_AUTH_KEY` | `tsnet.auth_key` |
| `SENTINEL_TAILSCALE_AUTH_KEY` | Tailscale onboarding auth key fallback used by runtime wiring |
| `SENTINEL_TSNET_CLIENT_SECRET` | `tsnet.client_secret` |
| `SENTINEL_TSNET_CLIENT_ID` | `tsnet.client_id` |
| `SENTINEL_TSNET_ID_TOKEN` | `tsnet.id_token` |
| `SENTINEL_TSNET_AUDIENCE` | `tsnet.audience` |
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

When both auth key and OAuth credentials are configured, Sentinel prioritizes auth key onboarding.
OAuth credentials are used when no auth key is resolved (or when `tsnet.login_mode=oauth`).

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
SENTINEL_TSNET_ADVERTISE_TAGS='["tag:sentinel"]' \
SENTINEL_TSNET_CLIENT_SECRET=oauth-client-secret \
SENTINEL_TSNET_CLIENT_ID=oauth-client-id \
SENTINEL_NOTIFIER_SINKS='[{"name":"stdout-debug","type":"stdout"},{"name":"discord-primary","type":"discord","url":"${SENTINEL_DISCORD_WEBHOOK_URL}"}]' \
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"device":{"names":["sentinel"],"tags":["tag:dev"],"owners":["123"],"ips":["100.64.0.10"]},"sinks":["stdout-debug","discord-primary"]}]' \
sentinel validate-config
```

## Routing Examples (JSON + Env)

All examples below assume sink names already exist under `notifier.sinks`.

Tips:
- `SENTINEL_NOTIFIER_ROUTES` must be a JSON array.
- `device` selectors are optional. If omitted, routing is based on `event_types`/`severities` only.
- Selector behavior is OR within a field and AND across fields.
- `device.owners` matches stable owner identity values (for example user IDs).

### 1) Only route notifications for devices with tag `tag:foo`

JSON file snippet (`sentinel.json`):

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["*"],
        "device": {
          "tags": ["tag:foo"]
        },
        "sinks": ["discord-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"device":{"tags":["tag:foo"]},"sinks":["discord-primary"]}]'
```

### 2) Only route `peer.online` events for devices with tag `tag:foo`

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["peer.online"],
        "device": {
          "tags": ["tag:foo"]
        },
        "sinks": ["discord-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["peer.online"],"device":{"tags":["tag:foo"]},"sinks":["discord-primary"]}]'
```

### 3) Only route online/offline events for one device name

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["peer.online", "peer.offline"],
        "device": {
          "names": ["laptop-01"]
        },
        "sinks": ["stdout-debug"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["peer.online","peer.offline"],"device":{"names":["laptop-01"]},"sinks":["stdout-debug"]}]'
```

### 4) Only route notifications for devices owned by user `123`

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["*"],
        "device": {
          "owners": ["123"]
        },
        "sinks": ["webhook-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"device":{"owners":["123"]},"sinks":["webhook-primary"]}]'
```

### 5) Only route notifications for devices with a specific IP

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["*"],
        "device": {
          "ips": ["100.64.0.10"]
        },
        "sinks": ["discord-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"device":{"ips":["100.64.0.10"]},"sinks":["discord-primary"]}]'
```

### 6) Only route online events for one exact device (`name` + `tag`)

This matches only when both conditions are true.

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["peer.online"],
        "device": {
          "names": ["node-a"],
          "tags": ["tag:prod"]
        },
        "sinks": ["webhook-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["peer.online"],"device":{"names":["node-a"],"tags":["tag:prod"]},"sinks":["webhook-primary"]}]'
```

### 7) Wildcard event families, but only for selected device subset

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["*"],
        "device": {
          "tags": ["tag:prod", "tag:staging"],
          "owners": ["123", "456"]
        },
        "sinks": ["stdout-debug", "discord-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"device":{"tags":["tag:prod","tag:staging"],"owners":["123","456"]},"sinks":["stdout-debug","discord-primary"]}]'
```

### 8) Multiple routes: targeted online alerts + broad fallback

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["peer.online"],
        "device": {
          "tags": ["tag:critical"]
        },
        "sinks": ["discord-primary"]
      },
      {
        "event_types": ["*"],
        "sinks": ["stdout-debug"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["peer.online"],"device":{"tags":["tag:critical"]},"sinks":["discord-primary"]},{"event_types":["*"],"sinks":["stdout-debug"]}]'
```

## Full Example

See [`config.example.yaml`](../config.example.yaml).
