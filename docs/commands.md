# Command Reference

Examples in this page use the installed `sentinel` binary.
For source-based development examples, see the development section in [Getting Started](getting-started.md).

## Core Commands

- `run`: start continuous observation and notification loop
- `status`: show current Sentinel + enrollment status
- `diff`: run one diff cycle and print results
- `dump-netmap`: print normalized netmap payload
- `test-notify`: send synthetic notification through notifier pipeline
- `validate-config`: validate merged runtime config

## Common Flags

- `--config`: path to YAML/JSON config
- `--log-format`: `pretty|json`
- `--log-level`: zap level string (`debug`, `info`, ...)
- `--no-color`: disable ANSI styling

## Tailscale Flags

- `--tailscale-login-mode`
- `--tailscale-auth-key`
- `--tailscale-state-dir`
- `--tailscale-login-timeout`
- `--tailscale-allow-interactive-fallback`

## Example Invocations

```bash
sentinel status --config ./config.example.yaml
```

```bash
sentinel run --config ./config.example.yaml --log-format json --log-level debug
```

```bash
sentinel diff --config ./config.example.yaml
```

```bash
sentinel test-notify --config ./config.example.yaml --dry-run
```

## Docker Command Example

```bash
docker run --rm \
  -e SENTINEL_TAILSCALE_AUTH_KEY=tskey-... \
  -e SENTINEL_CONFIG_PATH=/sentinel/config.yaml \
  -v "$(pwd)/config.example.yaml:/sentinel/config.yaml:ro" \
  ghcr.io/jaxxstorm/sentinel:latest status
```
