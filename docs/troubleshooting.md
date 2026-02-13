# Troubleshooting

## Event appears in stdout sink but not webhook

1. Confirm webhook sink URL resolved from environment.
2. Look for sink logs:
   - `webhook send succeeded`
   - `webhook send failed`
3. Verify endpoint availability independently with `curl`.
4. Keep process running long enough to observe retry completion.

## Repeated transitions and idempotency

Sentinel stores idempotency keys in the state file.

- State path defaults to `.sentinel/state.json`.
- Duplicate suppression applies when the same idempotency key is seen inside TTL.
- Idempotency keys are derived from event attributes including timestamp, so repeated real transitions at different times are delivered.

Inspect current state:

```bash
cat .sentinel/state.json
```

## Config validation failures

Run:

```bash
go run ./cmd/sentinel validate-config --config ./config.example.yaml
```

Common issues:

- invalid durations
- missing `state.path` or `tsnet.state_dir`
- unsupported `source.mode` or `tsnet.login_mode`
- unknown `notifier.routes[].event_types` values (use a documented event type or `*`)

## Wildcard routing checks

If you expect all events to be delivered, verify route config contains wildcard:

```yaml
notifier:
  routes:
    - event_types: ["*"]
      sinks: ["stdout-debug", "webhook-primary"]
```

If `event_types` omits `*`, only explicitly listed event types are delivered.

## Migrating from presence-only routes

Previous configs commonly used:

```yaml
event_types: ["peer.online", "peer.offline"]
```

To include expanded event families, either:

1. Switch to wildcard:
   - `event_types: ["*"]`
2. Keep explicit control and add selected types:
   - `peer.routes.changed`
   - `daemon.state.changed`
   - `prefs.run_ssh.changed`

## Local verification workflow

```bash
go run ./cmd/sentinel validate-config --config ./config.example.yaml
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
go run ./cmd/sentinel test-notify --config ./config.example.yaml
```

```bash
go run ./cmd/sentinel test-notify --config ./config.example.yaml --dry-run
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
go run ./cmd/sentinel run --config ./config.example.yaml --log-level debug
```
