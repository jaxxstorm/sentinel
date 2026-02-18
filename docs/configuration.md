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
5. shorthand composite notifier env vars (append behavior, non-JSON)

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
  - optional `filters` object narrows peer/device-scoped events:
    - `filters.include.device_names`: match by device name (supports glob patterns like `*.mullvad.ts.net`)
    - `filters.include.tags`: match when any configured tag is present
    - `filters.include.ips`: match by literal IP or CIDR
    - `filters.include.events`: match specific event types (supports `*`)
    - `filters.exclude.device_names`, `filters.exclude.tags`, `filters.exclude.ips`, `filters.exclude.events`: suppress matching events
  - filter semantics are deterministic:
    - OR within each field list
    - AND across configured fields in each include/exclude block
    - exclude match takes precedence over include match
    - routes with filters do not match non-device events (for example `daemon.state.changed`)
  - legacy `device.names`, `device.tags`, and `device.ips` fields are still accepted and mapped to include filters for compatibility
  - legacy `device.owners` remains supported for owner-based filtering

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

### Shorthand composite overrides (non-JSON, append behavior)

For notifier sinks/routes, Sentinel also supports non-JSON env vars. These **append** entries to config/default routes and sinks rather than replacing the entire list.

| Env var | Value format | Effect |
| --- | --- | --- |
| `SENTINEL_NOTIFIER_SINK_NAME` | string | appends one sink definition (`name`) |
| `SENTINEL_NOTIFIER_SINK_TYPE` | string | appends one sink definition (`type`) |
| `SENTINEL_NOTIFIER_SINK_URL` | string | appends one sink definition (`url`) |
| `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES` | comma list | appends one route (`event_types`) |
| `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPE` | comma list | deprecated alias for `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES` |
| `SENTINEL_NOTIFIER_ROUTE_SEVERITIES` | comma list | appends one route (`severities`) |
| `SENTINEL_NOTIFIER_ROUTE_SINKS` | comma list | appends one route (`sinks`) |
| `SENTINEL_NOTIFIER_SINK` | comma list | deprecated alias for `SENTINEL_NOTIFIER_ROUTE_SINKS` |
| `SENTINEL_NOTIFIER_ROUTE_DEVICE_NAMES` | comma list | appends one route (`device.names`) |
| `SENTINEL_NOTIFIER_ROUTE_DEVICE_TAGS` | comma list | appends one route (`device.tags`) |
| `SENTINEL_NOTIFIER_ROUTE_DEVICE_OWNERS` | comma list | appends one route (`device.owners`) |
| `SENTINEL_NOTIFIER_ROUTE_DEVICE_IPS` | comma list | appends one route (`device.ips`) |
| `SENTINEL_NOTIFIER_ROUTE_FILTER_INCLUDE_DEVICE_NAMES` | comma list | appends one route (`filters.include.device_names`) |
| `SENTINEL_NOTIFIER_ROUTE_FILTER_INCLUDE_TAGS` | comma list | appends one route (`filters.include.tags`) |
| `SENTINEL_NOTIFIER_ROUTE_FILTER_INCLUDE_IPS` | comma list | appends one route (`filters.include.ips`) |
| `SENTINEL_NOTIFIER_ROUTE_FILTER_INCLUDE_EVENTS` | comma list | appends one route (`filters.include.events`) |
| `SENTINEL_NOTIFIER_ROUTE_FILTER_EXCLUDE_DEVICE_NAMES` | comma list | appends one route (`filters.exclude.device_names`) |
| `SENTINEL_NOTIFIER_ROUTE_FILTER_EXCLUDE_TAGS` | comma list | appends one route (`filters.exclude.tags`) |
| `SENTINEL_NOTIFIER_ROUTE_FILTER_EXCLUDE_IPS` | comma list | appends one route (`filters.exclude.ips`) |
| `SENTINEL_NOTIFIER_ROUTE_FILTER_EXCLUDE_EVENTS` | comma list | appends one route (`filters.exclude.events`) |

Notes:
- Scalar fields (for example `SENTINEL_POLL_INTERVAL`, `SENTINEL_TSNET_HOSTNAME`) are already settable directly without JSON.
- Use structured JSON env vars when you want full replacement control for complete sink/route arrays.
- Canonical shorthand keys are `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES` and `SENTINEL_NOTIFIER_ROUTE_SINKS`.
- If both canonical and deprecated alias keys are set, Sentinel uses canonical values.

### Migration Notes (Legacy Shorthand Aliases)

Use these canonical replacements:

| Deprecated key | Canonical key |
| --- | --- |
| `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPE` | `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES` |
| `SENTINEL_NOTIFIER_SINK` | `SENTINEL_NOTIFIER_ROUTE_SINKS` |

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
SENTINEL_NOTIFIER_SINK_NAME=discord-primary \
SENTINEL_NOTIFIER_SINK_TYPE=discord \
SENTINEL_NOTIFIER_SINK_URL='${SENTINEL_DISCORD_WEBHOOK_URL}' \
SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES='*' \
SENTINEL_NOTIFIER_ROUTE_FILTER_INCLUDE_EVENTS='*' \
SENTINEL_NOTIFIER_ROUTE_FILTER_INCLUDE_DEVICE_NAMES='sentinel' \
SENTINEL_NOTIFIER_ROUTE_FILTER_INCLUDE_TAGS='tag:dev' \
SENTINEL_NOTIFIER_ROUTE_FILTER_INCLUDE_IPS='100.64.0.10' \
SENTINEL_NOTIFIER_ROUTE_SINKS='stdout-debug,discord-primary' \
sentinel validate-config
```

## Routing Examples (JSON + Env)

All examples below assume sink names already exist under `notifier.sinks`.

Tips:
- `SENTINEL_NOTIFIER_ROUTES` must be a JSON array.
- `filters` are optional. If omitted, routing is based on `event_types`/`severities` only.
- Include/exclude behavior is OR within a field and AND across configured fields; exclude takes precedence.
- `filters.include.device_names` supports glob patterns (for example `*.mullvad.ts.net`).
- `filters.include.events` / `filters.exclude.events` support explicit event types and `*`.
- Legacy `device.names`, `device.tags`, and `device.ips` are accepted as compatibility aliases for include filters.

### Event Type Reference (all emitted event types)

- `peer.online`
- `peer.offline`
- `peer.added`
- `peer.removed`
- `peer.routes.changed`
- `peer.tags.changed`
- `peer.machine_authorized.changed`
- `peer.key_expiry.changed`
- `peer.key_expired`
- `peer.hostinfo.changed`
- `daemon.state.changed`
- `prefs.advertise_routes.changed`
- `prefs.exit_node.changed`
- `prefs.run_ssh.changed`
- `prefs.shields_up.changed`
- `tailnet.domain.changed`
- `tailnet.tka_enabled.changed`

### 1) Only route notifications for devices with tag `tag:foo`

JSON file snippet (`sentinel.json`):

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["*"],
        "filters": {
          "include": {
            "tags": ["tag:foo"]
          }
        },
        "sinks": ["discord-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"filters":{"include":{"tags":["tag:foo"]}},"sinks":["discord-primary"]}]'
```

### 2) Only route `peer.online` events for devices with tag `tag:foo`

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["peer.online"],
        "filters": {
          "include": {
            "tags": ["tag:foo"]
          }
        },
        "sinks": ["discord-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["peer.online"],"filters":{"include":{"tags":["tag:foo"]}},"sinks":["discord-primary"]}]'
```

### 3) Only route online/offline events for one device name

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["peer.online", "peer.offline"],
        "filters": {
          "include": {
            "device_names": ["laptop-01"]
          }
        },
        "sinks": ["stdout-debug"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["peer.online","peer.offline"],"filters":{"include":{"device_names":["laptop-01"]}},"sinks":["stdout-debug"]}]'
```

### 4) Only route notifications for devices with a specific IP (or CIDR)

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["*"],
        "filters": {
          "include": {
            "ips": ["100.64.0.0/24"]
          }
        },
        "sinks": ["discord-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"filters":{"include":{"ips":["100.64.0.0/24"]}},"sinks":["discord-primary"]}]'
```

### 5) Suppress Mullvad shared nodes while keeping other peer notifications

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["*"],
        "filters": {
          "exclude": {
            "device_names": ["*.mullvad.ts.net"]
          }
        },
        "sinks": ["stdout-debug"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"filters":{"exclude":{"device_names":["*.mullvad.ts.net"]}},"sinks":["stdout-debug"]}]'
```

### 6) Only route online events for one exact device (`device_name` + `tag`)

This matches only when both conditions are true.

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["peer.online"],
        "filters": {
          "include": {
            "device_names": ["node-a"],
            "tags": ["tag:prod"]
          }
        },
        "sinks": ["webhook-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["peer.online"],"filters":{"include":{"device_names":["node-a"],"tags":["tag:prod"]}},"sinks":["webhook-primary"]}]'
```

### 7) Wildcard event families, but only for selected device subset

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["*"],
        "filters": {
          "include": {
            "tags": ["tag:prod", "tag:staging"],
            "ips": ["100.64.0.0/24"]
          },
          "exclude": {
            "device_names": ["*.mullvad.ts.net"]
          }
        },
        "sinks": ["stdout-debug", "discord-primary"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"filters":{"include":{"tags":["tag:prod","tag:staging"],"ips":["100.64.0.0/24"]},"exclude":{"device_names":["*.mullvad.ts.net"]}},"sinks":["stdout-debug","discord-primary"]}]'
```

### 8) Multiple routes: targeted online alerts + broad fallback

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["peer.online"],
        "filters": {
          "include": {
            "tags": ["tag:critical"]
          }
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
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["peer.online"],"filters":{"include":{"tags":["tag:critical"]}},"sinks":["discord-primary"]},{"event_types":["*"],"sinks":["stdout-debug"]}]'
```

### 9) Include/exclude event filters with wildcard route

This keeps a wildcard route but suppresses non-actionable event families.

JSON file snippet:

```json
{
  "notifier": {
    "routes": [
      {
        "event_types": ["*"],
        "filters": {
          "include": {
            "events": ["*"]
          },
          "exclude": {
            "events": ["peer.routes.changed", "peer.tags.changed"]
          }
        },
        "sinks": ["stdout-debug"]
      }
    ]
  }
}
```

Env var equivalent:

```bash
SENTINEL_NOTIFIER_ROUTES='[{"event_types":["*"],"filters":{"include":{"events":["*"]},"exclude":{"events":["peer.routes.changed","peer.tags.changed"]}},"sinks":["stdout-debug"]}]'
```

## Full Example

See [`config.example.yaml`](../config.example.yaml).
