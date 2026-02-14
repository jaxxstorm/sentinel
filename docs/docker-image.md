# Docker Image

This page documents environment variables used when running Sentinel from the published container image.

## Image Defaults

- Image entrypoint: `sentinel`
- Default command: `run`
- Default config path in image: `/sentinel/config.yaml`
- Default environment in image:
  - `SENTINEL_CONFIG_PATH=/sentinel/config.yaml`

## Runtime Environment Variables

| Variable | Required | Purpose |
| --- | --- | --- |
| `SENTINEL_CONFIG_PATH` | No | Config file path used by Sentinel. In the image this defaults to `/sentinel/config.yaml`. |
| `SENTINEL_TAILSCALE_AUTH_KEY` | Depends on login mode | Auth key for Tailscale onboarding (`tsnet.login_mode=auth_key` requires this unless `tsnet.auth_key` is set in config). |
| `SENTINEL_TSNET_AUTH_KEY` | No | Maps to `tsnet.auth_key` in config. Used as a lower-priority auth key source than `SENTINEL_TAILSCALE_AUTH_KEY`. |
| `SENTINEL_TSNET_ADVERTISE_TAGS` | No | Tags requested during enrollment. Accepts JSON array or comma-separated tags (for example `tag:sentinel,tag:prod`). |
| `SENTINEL_TSNET_CLIENT_SECRET` | Depends on login mode | OAuth client secret for tsnet onboarding credentials. |
| `SENTINEL_TSNET_CLIENT_ID` | Required with client secret | OAuth client identifier paired with `SENTINEL_TSNET_CLIENT_SECRET`. |
| `SENTINEL_TSNET_ID_TOKEN` | No | Optional tsnet identity token input for OAuth-related flows. |
| `SENTINEL_TSNET_AUDIENCE` | No | Optional tsnet audience input for OAuth-related flows. |
| `SENTINEL_STATE_PATH` | No | Overrides `state.path` for idempotency/state storage. Useful for mounting persistent state in containers. |
| `SENTINEL_TSNET_STATE_DIR` | Recommended | Sets `tsnet.state_dir` for persisted tsnet state in container volumes. |

Sentinel also supports general config overrides using `SENTINEL_` variables mapped from config keys. Example:

- `SENTINEL_OUTPUT_LOG_FORMAT=json`
- `SENTINEL_OUTPUT_LOG_LEVEL=debug`
- `SENTINEL_TSNET_STATE_DIR=/var/lib/sentinel/tsnet`
- `SENTINEL_TSNET_LOGIN_MODE=oauth`

When both auth key and OAuth credentials are configured, Sentinel prioritizes auth key onboarding.

## Structured Environment Variables (JSON)

These keys are used for complex sections that are hard to express with scalar env vars:

| Variable | JSON type | Example |
| --- | --- | --- |
| `SENTINEL_DETECTORS` | object | `{"presence":{"enabled":true},"runtime":{"enabled":true}}` |
| `SENTINEL_DETECTOR_ORDER` | array | `["presence","peer_changes","runtime"]` |
| `SENTINEL_NOTIFIER_SINKS` | array | `[{"name":"stdout-debug","type":"stdout"}]` |
| `SENTINEL_NOTIFIER_ROUTES` | array | `[{"event_types":["*"],"sinks":["stdout-debug"]}]` |

If these keys are present, they override their entire corresponding config sections.

## Placeholder Environment Variables in Config

If your config uses `${...}` placeholders for sink URLs, you must provide those environment variables at runtime.

Common examples:

- `SENTINEL_WEBHOOK_URL`
- `SENTINEL_DISCORD_WEBHOOK_URL`
- `REQUESTBIN_WEBHOOK_URL`

Example config snippet:

```yaml
notifier:
  sinks:
    - name: webhook-primary
      type: webhook
      url: ${SENTINEL_WEBHOOK_URL}
    - name: discord-primary
      type: discord
      url: ${SENTINEL_DISCORD_WEBHOOK_URL}
```

## Example `docker run`

```bash
docker run --rm \
  -e SENTINEL_CONFIG_PATH=/sentinel/config.yaml \
  -e SENTINEL_TAILSCALE_AUTH_KEY=tskey-... \
  -e SENTINEL_TSNET_ADVERTISE_TAGS='["tag:sentinel"]' \
  -e SENTINEL_TSNET_CLIENT_SECRET=oauth-client-secret \
  -e SENTINEL_TSNET_CLIENT_ID=oauth-client-id \
  -e SENTINEL_WEBHOOK_URL=https://example.test/webhook \
  -e SENTINEL_DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... \
  -v "$(pwd)/config.example.yaml:/sentinel/config.yaml:ro" \
  -v "$(pwd)/.sentinel:/var/lib/sentinel" \
  ghcr.io/<owner>/<repo>:latest run
```

If you mount state at `/var/lib/sentinel`, set:

```bash
-e SENTINEL_STATE_PATH=/var/lib/sentinel/state.json
-e SENTINEL_TSNET_STATE_DIR=/var/lib/sentinel/tsnet
```

## Environment-Only `docker run` (No Mounted Config)

This mode avoids mounting `config.yaml` entirely:

```bash
docker run --rm \
  -e SENTINEL_STATE_PATH=/var/lib/sentinel/state.json \
  -e SENTINEL_TSNET_STATE_DIR=/var/lib/sentinel/tsnet \
  -e SENTINEL_TSNET_LOGIN_MODE=auto \
  -e SENTINEL_TAILSCALE_AUTH_KEY=tskey-... \
  -e SENTINEL_TSNET_ADVERTISE_TAGS='["tag:sentinel"]' \
  -e SENTINEL_TSNET_CLIENT_SECRET=oauth-client-secret \
  -e SENTINEL_TSNET_CLIENT_ID=oauth-client-id \
  -e SENTINEL_DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... \
  -e SENTINEL_NOTIFIER_SINKS='[{\"name\":\"stdout-debug\",\"type\":\"stdout\"},{\"name\":\"discord-primary\",\"type\":\"discord\",\"url\":\"${SENTINEL_DISCORD_WEBHOOK_URL}\"}]' \
  -e SENTINEL_NOTIFIER_ROUTES='[{\"event_types\":[\"*\"],\"sinks\":[\"stdout-debug\",\"discord-primary\"]}]' \
  -v "$(pwd)/.sentinel:/var/lib/sentinel" \
  ghcr.io/<owner>/<repo>:latest run
```
