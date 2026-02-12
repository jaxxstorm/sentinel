## 1. Docsify Site Structure

- [x] 1.1 Create `docs/` scaffold for Docsify rendering (`docs/README.md`, navigation/sidebar files, and supporting pages).
- [x] 1.2 Add a docs landing page describing Sentinel purpose, primary workflows, and links to configuration and sink guides.
- [x] 1.3 Ensure docs navigation is consistent and all top-level docs pages are reachable from the sidebar.

## 2. Configuration and Sink Documentation

- [x] 2.1 Write a configuration reference page covering each top-level config section, field intent, defaults, and expected value types.
- [x] 2.2 Document sink configuration and routing behavior, including `stdout/debug`, webhook sinks, route matching, and fallback behavior.
- [x] 2.3 Add environment variable documentation for both `SENTINEL_` overrides and `${VAR_NAME}` interpolation in sink URLs.
- [x] 2.4 Add runnable examples that align with current `config.example.yaml` and current runtime logging/delivery behavior.

## 3. Troubleshooting and Verification Guides

- [x] 3.1 Add troubleshooting guidance for missing webhook deliveries, retry behavior, and interpreting sink success/failure logs.
- [x] 3.2 Document idempotency suppression behavior and how to verify whether an event was suppressed versus delivered.
- [x] 3.3 Add a local validation workflow using `validate-config`, `test-notify`, and `run --dry-run`.

## 4. README Rewrite

- [x] 4.1 Rewrite root `README.md` in a concise conventional format (overview, quick start, key commands, docs links) without emoji styling.
- [x] 4.2 Remove redundant long-form detail from README and defer operational depth to `docs/` pages.
- [x] 4.3 Ensure README examples and command snippets are accurate for current CLI and config behavior.

## 5. Quality Checks

- [x] 5.1 Verify all new documentation links resolve and cross-references between README and docs are correct.
- [x] 5.2 Perform a docsify local render check to confirm pages and sidebar render as expected.
- [x] 5.3 Review text for plain technical tone and remove generated-sounding phrasing.
