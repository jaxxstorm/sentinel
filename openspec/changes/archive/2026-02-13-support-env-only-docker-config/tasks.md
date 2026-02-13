## 1. Structured Environment Parsing

- [x] 1.1 Add config parser support for structured environment keys (`SENTINEL_DETECTORS`, `SENTINEL_DETECTOR_ORDER`, `SENTINEL_NOTIFIER_SINKS`, `SENTINEL_NOTIFIER_ROUTES`).
- [x] 1.2 Define and enforce deterministic precedence between defaults, file config, scalar env overrides, and structured env keys.
- [x] 1.3 Add actionable parse errors that include offending environment key names.

## 2. Environment-Only Runtime Support

- [x] 2.1 Ensure Sentinel can start with no config file when required settings are present via env vars.
- [x] 2.2 Preserve backward compatibility for YAML/JSON loading and existing `${VAR}` sink URL interpolation.
- [x] 2.3 Validate mixed-mode behavior (file + scalar env + structured env) for deterministic outcomes.

## 3. Validation and Test Coverage

- [x] 3.1 Add unit tests for valid structured env parsing of detectors, detector order, sinks, and routes.
- [x] 3.2 Add unit tests for malformed structured env values and key-attributed error messages.
- [x] 3.3 Add regression tests confirming existing file-based config behavior remains unchanged.

## 4. Documentation for Docker/Platform Deployments

- [x] 4.1 Update docs with a complete env-var matrix for scalar and structured config options.
- [x] 4.2 Add env-only `docker run -e` and platform deployment examples that avoid mounted config files.
- [x] 4.3 Add troubleshooting guidance for env parsing, precedence, and validation failures.

## 5. Verification

- [x] 5.1 Run config-focused test suites and full `go test ./...` to validate parser and runtime behavior.
- [x] 5.2 Verify `validate-config` and `run` behavior using environment-only sample configurations.
