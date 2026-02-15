## 1. Compose templates and environment surface

- [x] 1.1 Add a Railway-importable base `docker-compose.yml` that runs Sentinel from GHCR image defaults.
- [x] 1.2 Add a local override compose file (for example `docker-compose.local.yml`) that builds the image from local source and preserves compatible runtime settings.
- [x] 1.3 Add/update `.env.example` and related ignore rules so all supported `SENTINEL_` runtime variables are represented with clear required vs optional guidance and no live secrets.

## 2. Release image tagging behavior

- [x] 2.1 Update Docker publish workflow behavior so compose template defaults can rely on a maintained `latest` GHCR tag alongside versioned tags.
- [x] 2.2 Ensure release validation logic/docs reflect the updated tagging behavior and expected publish triggers.

## 3. Operator documentation

- [x] 3.1 Update compose/containers documentation with local run instructions and Railway drag-and-drop template usage.
- [x] 3.2 Add an explicit environment variable matrix in docs that maps compose variables to required/optional semantics and mode-specific auth expectations.
- [x] 3.3 Document secret handling practices for local `.env` and Railway-managed variable storage.

## 4. Validation

- [x] 4.1 Validate compose files with `docker compose config` for base and local overlay paths.
- [x] 4.2 Validate local build/run flow using compose and confirm Sentinel starts with documented required inputs.
- [x] 4.3 Run relevant workflow/config checks (including release validation checks) and capture outcomes in the change notes/PR summary.
