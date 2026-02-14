## 1. Configuration and validation surface

- [x] 1.1 Extend `TSNetConfig` with advertise tag and OAuth credential fields plus defaults.
- [x] 1.2 Add config/env mappings for new tsnet fields (including env-only container paths) and preserve existing auth-key precedence.
- [x] 1.3 Implement validation for advertise tag format and OAuth credential field combinations with actionable errors.

## 2. Onboarding credential and tsnet wiring

- [x] 2.1 Add onboarding credential resolution for OAuth alongside existing auth-key sources.
- [x] 2.2 Update onboarding mode selection to support `existing state > auth key > OAuth > interactive` (or explicit OAuth mode if selected).
- [x] 2.3 Apply configured `AdvertiseTags` and OAuth credential fields to `tsnet.Server` before first status probe/start.
- [x] 2.4 Extend onboarding/status output to report credential mode/source without exposing secret material.

## 3. Tests and operator documentation

- [x] 3.1 Add unit tests for config parsing and validation of advertise tags and OAuth credential combinations.
- [x] 3.2 Add onboarding tests for credential precedence (auth key vs OAuth), OAuth path selection, and redaction safety.
- [x] 3.3 Update docs and examples (`docs/configuration.md`, `docs/docker-image.md`, `config.example.yaml`) for tags and OAuth credential usage.
- [x] 3.4 Run focused test suites and `validate-config` checks for new onboarding scenarios.
