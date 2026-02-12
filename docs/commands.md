# Command Reference

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
go run ./cmd/sentinel status --config ./config.example.yaml
```

```bash
go run ./cmd/sentinel run --config ./config.example.yaml --log-format json --log-level debug
```

```bash
go run ./cmd/sentinel diff --config ./config.example.yaml
```
