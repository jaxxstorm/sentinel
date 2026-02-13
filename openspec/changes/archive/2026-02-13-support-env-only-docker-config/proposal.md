## Why

Running Sentinel on container-first platforms (for example Railway) should not require mounting a config file. Today, some settings can be overridden with `SENTINEL_` env vars, but complex nested config is still effectively file-only, which blocks pure env-var deployments.

## What Changes

- Add a full env-driven configuration ingestion path that can express all Sentinel config options, including structured fields such as sink and route lists.
- Define deterministic env encoding rules for nested objects, arrays, durations, booleans, and strings so container-only deployments are predictable.
- Preserve backward compatibility with existing YAML/JSON file loading and current `${VAR}` sink URL interpolation behavior.
- Add validation and operator-facing error messages for malformed env-encoded structured config.
- Expand docs with a complete env-var reference and docker-first deployment examples using only `docker run -e` or platform env settings.

## Capabilities

### New Capabilities
- `env-only-config-ingestion`: Sentinel accepts a complete runtime configuration from environment variables only, including complex notifier and detector structures.

### Modified Capabilities
- `sentinel-cli-config`: broaden env override requirements from partial overrides to full config coverage and deterministic structured parsing behavior.
- `docsify-operator-guides`: document full env-var configuration patterns for docker/platform deployments, including complex list/object examples and troubleshooting.

## Impact

- Affected code: `internal/config` loading/parsing/validation logic, config tests, and wiring assumptions that currently depend on file-oriented structures.
- Affected docs: `docs/configuration.md`, `docs/docker-image.md`, `docs/troubleshooting.md`, and related examples.
- Runtime behavior: Sentinel can boot with `SENTINEL_CONFIG_PATH` unset when sufficient env vars are provided.
- Deployment systems: docker, Railway, and similar platforms can run Sentinel with env vars only and no mounted config file.
