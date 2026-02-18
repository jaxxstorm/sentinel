## 1. Config model and validation

- [x] 1.1 Extend notifier route config structs with `filters.include` and `filters.exclude` fields for `device_names`, `tags`, `ips`, and `events`.
- [x] 1.2 Implement compatibility mapping from legacy `notifier.routes[].device.*` selectors to include filters.
- [x] 1.3 Add validation for filter values (non-empty names/tags, valid IP/CIDR entries) and conflict handling.
- [x] 1.4 Add config unit tests for valid include/exclude filters (including `events`), invalid values, and legacy selector compatibility.

## 2. Route matching implementation

- [x] 2.1 Extend route matching inputs to provide normalized device-name, tag, and IP identity fields for filter evaluation.
- [x] 2.2 Implement include filter matching semantics (OR within field values, AND across configured filter fields including `events`).
- [x] 2.3 Implement exclude filter semantics with precedence over include matches.
- [x] 2.4 Add matcher unit tests for device-name glob matching (including `*.mullvad.ts.net`), tag filters, IP/CIDR filters, event filters, and precedence behavior.

## 3. Operator docs and regression verification

- [x] 3.1 Update configuration docs to describe the global notification filter system and legacy selector compatibility.
- [x] 3.2 Add concrete examples for filtering by device name, tag, IP, and event type, including Mullvad suppression via exclude filters.
- [x] 3.3 Run notifier/config test suites and verify routes without filters preserve current behavior.
