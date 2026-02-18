## Why

Notification routing currently lacks a unified filter system, which makes it difficult to reduce noise without losing useful events. We need global notification filters so operators can target or suppress events by device identity attributes, including common noisy cases such as Mullvad shared nodes.

## What Changes

- Introduce a global notification filter system on notifier routes for device names, IPs, tags, and event types.
- Support include and exclude filter semantics so routes can suppress noisy subsets without disabling broader event families.
- Preserve backward compatibility by mapping existing device selector behavior into the new filter model.
- Document filter-based routing with a Mullvad suppression example as one practical use case.

## Capabilities

### New Capabilities
- `notifier-notification-filter-system`: Route matching supports global notification filters for device names, IPs, tags, and event types with include/exclude semantics.

### Modified Capabilities
- `notification-delivery-pipeline`: Route matching behavior changes to apply global include/exclude filters during sink routing decisions.
- `sentinel-cli-config`: Config schema/validation expands to accept notification filter fields and documented filter examples.
- `notifier-device-target-filtering`: Existing selector behavior is retained through compatibility mapping into the new filter system.

## Impact

- Affected code: notifier route matcher, event identity extraction, config model/validation, and configuration documentation.
- API/config impact: new optional `filters` route fields with include/exclude support for device names, tags, IPs, and event types; existing selector configuration remains backward-compatible.
- Operational impact: operators can implement broad notification routing while suppressing noisy targets (including Mullvad patterns) via explicit filters.
