## Context

Sentinel currently supports route filtering by event type, severity, and optional device selectors (`names`, `tags`, `owners`, `ips`), but those selectors are not presented as a unified global filter system. Operators need a consistent way to filter notifications by device identity attributes (device name, IP, tag) across all routes, including suppression of noisy classes such as Mullvad-hosted nodes.

## Goals / Non-Goals

**Goals:**
- Add a unified notification filter system that supports device-name, IP, tag, and event-type filters for every notifier route.
- Support include and exclude behavior so noisy subsets can be suppressed from otherwise broad routing rules.
- Allow Mullvad suppression as one use case through device-name pattern matching.
- Preserve backward compatibility for existing route configuration that does not define peer filters.
- Keep filter evaluation deterministic and composable with existing route predicates.
- Provide clear operator documentation and config validation for the new filter system.

**Non-Goals:**
- Redesigning notifier routing architecture beyond route predicate evaluation.
- Implementing arbitrary regex-based filtering for all fields.
- Adding sink-specific behavior differences for filters.

## Decisions

### Add `notifier.routes[].filters` as the canonical filter namespace
- Decision: Introduce a route-level `filters` object with optional `include` and `exclude` blocks, each supporting `device_names`, `tags`, `ips`, and `events`.
- Rationale: A dedicated namespace makes filtering explicit and extensible while providing a consistent operator experience.
- Alternative considered: Keeping only legacy `device` selector naming. Rejected because it does not communicate global filter semantics.

### Keep compatibility with existing device selector fields
- Decision: Preserve `notifier.routes[].device.*` config as backward-compatible input and map it to equivalent `filters.include` behavior at load time.
- Rationale: Avoids breaking existing deployments while moving to a clearer filter model.
- Alternative considered: Hard cutover to `filters` only. Rejected due to migration risk.

### Evaluate filters with deterministic route semantics
- Decision: Keep existing route predicate combination rules and apply filters in the same matching phase:
  - OR within each list
  - AND across configured filter dimensions
  - exclude filters take precedence over include filters
  - route must satisfy event type, severity, and filter predicates
- Rationale: Matches existing operator mental model and avoids surprises.
- Alternative considered: Priority-based filter ordering or early exclusions. Rejected because it complicates debugging and predictability.

### Validate and normalize filters in config loading
- Decision: Extend config validation to reject empty/invalid filter entries and normalize values for runtime matching (`device_names` glob, `ips` literal/CIDR, `tags` canonical values, and known event types including wildcard `*`).
- Rationale: Failing fast in config validation prevents runtime silent misrouting.
- Alternative considered: Best-effort runtime parsing. Rejected due to harder operator troubleshooting.

## Risks / Trade-offs

- [Risk] Running legacy selector mapping and new filters together can create ambiguous configuration.
  -> Mitigation: Define precedence rules and emit clear validation or warning output for conflicting fields.
- [Risk] Broad include/exclude patterns can over-filter legitimate notifications.
  -> Mitigation: Recommend dry-run verification and include matching diagnostics in route decision logs.
- [Risk] Additional predicate checks increase routing cost.
  -> Mitigation: Keep evaluation O(patterns) per route and reuse existing matcher pipeline.
