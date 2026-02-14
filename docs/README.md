# Sentinel Documentation

Sentinel runs as a tsnet-embedded Tailscale node, watches tailnet netmap changes, and emits notifications when policy-relevant events occur.

Use this documentation for day-to-day operation:

- [Getting Started](getting-started.md)
- [Configuration Reference](configuration.md)
- [Docker Compose and Railway](docker-compose.md)
- [Docker Image](docker-image.md)
- [Release Artifacts](release-artifacts.md)
- [Sinks and Routing](sinks-and-routing.md)
- [Troubleshooting](troubleshooting.md)
- [Command Reference](commands.md)

## Primary Workflow

1. Configure Sentinel in `config.example.yaml` (or your own config file).
2. Validate configuration.
3. Run one-shot notification checks (`test-notify`, `--dry-run`).
4. Run Sentinel continuously with `run`.
5. Use sink logs plus webhook responses to verify delivery.

## Runtime Notes

- Default `source.mode` is `realtime` (IPNBus stream).
- Default notifier behavior includes a local `stdout-debug` sink.
- Route `event_types` supports wildcard `*` to match all emitted event families.
- Runtime logs include `log_source` values: `sentinel`, `tailscale`, and `sink`.
