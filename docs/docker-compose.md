# Docker Compose and Railway Template

Sentinel ships a Compose-first deployment model:

- `docker-compose.yml`: canonical base template for Railway import and GHCR runtime images.
- `docker-compose.local.yml`: local development overlay that builds Sentinel from source.
- `.env.example`: complete runtime variable template with required vs optional guidance.

## Quick Start (Local)

1. Copy environment template:
   ```bash
   cp .env.example .env
   ```
2. Set required auth value (default mode is `auth_key`):
   ```bash
   SENTINEL_TAILSCALE_AUTH_KEY=tskey-auth-...   # in .env
   ```
3. Run locally with source build overlay:
   ```bash
   docker compose -f docker-compose.yml -f docker-compose.local.yml up --build
   ```

State data is stored in `./.sentinel` via the local overlay volume mount.

## Railway Template Import

1. In Railway, choose **New Project** -> **Deploy from Template / Docker Compose**.
2. Import `docker-compose.yml` from this repository.
3. Configure required variables in Railway:
   - `SENTINEL_TAILSCALE_AUTH_KEY` (required for default `auth_key` mode)
4. Optionally configure any additional `SENTINEL_` variables from the matrix below.

The default image in the template is `ghcr.io/jaxxstorm/sentinel:latest`. Pin to a version tag (for example `ghcr.io/jaxxstorm/sentinel:v1.2.3`) when deterministic rollout behavior is required.

## Compose Environment Matrix

### Required by default template behavior

| Variable | Required | Notes |
| --- | --- | --- |
| `SENTINEL_TAILSCALE_AUTH_KEY` | Yes (default) | Required when `SENTINEL_TSNET_LOGIN_MODE=auth_key` (template default). |

### Conditionally required by login mode

| Variable | Required | Notes |
| --- | --- | --- |
| `SENTINEL_TSNET_CLIENT_SECRET` | Required for `oauth` mode | Required when `SENTINEL_TSNET_LOGIN_MODE=oauth`. |
| `SENTINEL_TSNET_CLIENT_ID` | Required with client secret | Must be set when `SENTINEL_TSNET_CLIENT_SECRET` is set. |

### Optional runtime overrides

| Variable | Required | Notes |
| --- | --- | --- |
| `SENTINEL_IMAGE` | No | Compose image override (for example pinned version tag). |
| `SENTINEL_CONFIG_PATH` | No | Leave blank for env-only mode. |
| `SENTINEL_STATE_PATH` | No | Defaults to `/data/state.json` in template. |
| `SENTINEL_TSNET_STATE_DIR` | No | Defaults to `/data/tsnet` in template. |
| `SENTINEL_POLL_INTERVAL` | No | Maps to `poll_interval`. |
| `SENTINEL_POLL_JITTER` | No | Maps to `poll_jitter`. |
| `SENTINEL_POLL_BACKOFF_MIN` | No | Maps to `poll_backoff_min`. |
| `SENTINEL_POLL_BACKOFF_MAX` | No | Maps to `poll_backoff_max`. |
| `SENTINEL_SOURCE_MODE` | No | `realtime` or `poll`. |
| `SENTINEL_DETECTORS` | No | Structured JSON object override. |
| `SENTINEL_DETECTOR_ORDER` | No | Structured JSON array override. |
| `SENTINEL_POLICY_DEBOUNCE_WINDOW` | No | Maps to `policy.debounce_window`. |
| `SENTINEL_POLICY_SUPPRESSION_WINDOW` | No | Maps to `policy.suppression_window`. |
| `SENTINEL_POLICY_RATE_LIMIT_PER_MIN` | No | Maps to `policy.rate_limit_per_min`. |
| `SENTINEL_POLICY_BATCH_SIZE` | No | Maps to `policy.batch_size`. |
| `SENTINEL_NOTIFIER_SINKS` | No | Structured JSON array override. |
| `SENTINEL_NOTIFIER_ROUTES` | No | Structured JSON array override. |
| `SENTINEL_WEBHOOK_URL` | No | Used by `${SENTINEL_WEBHOOK_URL}` placeholders. |
| `SENTINEL_DISCORD_WEBHOOK_URL` | No | Used by `${SENTINEL_DISCORD_WEBHOOK_URL}` placeholders. |
| `SENTINEL_OUTPUT_LOG_FORMAT` | No | `pretty` or `json`. |
| `SENTINEL_OUTPUT_LOG_LEVEL` | No | Log level override. |
| `SENTINEL_OUTPUT_NO_COLOR` | No | Color output toggle. |
| `SENTINEL_TSNET_LOGIN_MODE` | No | `auto`, `auth_key`, `oauth`, `interactive`. |
| `SENTINEL_TSNET_AUTH_KEY` | No | Lower-priority auth key source than `SENTINEL_TAILSCALE_AUTH_KEY`. |
| `SENTINEL_TSNET_HOSTNAME` | No | Maps to `tsnet.hostname`. |
| `SENTINEL_TSNET_ADVERTISE_TAGS` | No | JSON array or comma-separated tags. |
| `SENTINEL_TSNET_ID_TOKEN` | No | Optional OAuth companion field. |
| `SENTINEL_TSNET_AUDIENCE` | No | Optional OAuth companion field. |
| `SENTINEL_TSNET_ALLOW_INTERACTIVE_FALLBACK` | No | Fallback behavior switch. |
| `SENTINEL_TSNET_LOGIN_TIMEOUT` | No | Interactive login timeout override. |

## Secret Handling

Sentinel templates intentionally keep secrets externalized:

- Do not commit `.env` files with real values.
- Keep `.env.example` placeholders only.
- Set sensitive values through:
  - local `.env` (git-ignored), or
  - Railway variable/secret management.

Never place live webhook URLs, auth keys, or OAuth client secrets directly in committed compose YAML.
