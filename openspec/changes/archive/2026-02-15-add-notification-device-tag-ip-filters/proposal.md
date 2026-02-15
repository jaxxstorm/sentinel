## Why

Sentinel currently routes notifications for every matching event type across the whole tailnet, which creates noise when operators only care about specific devices or groups. We need first-class target filters now so operators can scope notifications by device name, tag, owner, or IP without reducing event coverage to only `peer.online`/`peer.offline`.

## What Changes

- Add route-level device target filters so notifier routes can select devices by name, tags, owners, and IPs.
- Define deterministic matching semantics for combining `event_types` and device filters, including behavior for non-device events.
- Expand normalized device identity data to carry owner and IP values required for route filtering in both poll and realtime source modes.
- Extend device-scoped event payload baselines to include selector-ready identity fields so routing and sinks share stable context.
- Update config validation and examples to support device target filters in file-based and structured env-based configuration.

## Capabilities

### New Capabilities
- `notifier-device-target-filtering`: Add device-scoped route filters (name, tag, owner, IP) with deterministic matching semantics.

### Modified Capabilities
- `sentinel-cli-config`: Validate and load device target filter fields for notifier routes from YAML/JSON and structured env config.
- `notification-delivery-pipeline`: Apply device target filters during route matching while preserving existing behavior for routes without filters.
- `netmap-presence-diffing`: Include stable device owner and IP identity fields in normalized snapshots used by poll-mode diffing.
- `realtime-netmap-diff-pipeline`: Include stable device owner and IP identity fields in normalized snapshots derived from IPNBus realtime updates.
- `ipnbus-event-catalog`: Ensure device-scoped events include stable identity fields (name, tags, owners, IPs) for downstream routing and sink context.

## Impact

- Affected code:
  - `internal/config` (route schema, validation, env decoding)
  - `internal/notify` (route matching/filter evaluation)
  - `internal/source`, `internal/snapshot`, `internal/diff`, `internal/event` (device identity normalization and event payload fields)
  - `config.example.yaml` and docs (`docs/configuration.md`) for operator guidance
- External behavior:
  - Operators can scope notifications to selected devices or device groups without reducing event-type coverage.
  - Existing route definitions remain valid and keep current behavior when device filters are not configured.
