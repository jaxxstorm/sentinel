## Why

Sentinel now has enough functionality that setup and operations are difficult without consolidated documentation. We need a structured docs site for operators and a concise root README that reflects the real workflow and avoids generated-looking boilerplate.

## What Changes

- Add a docs site under `docs/` structured for Docsify rendering.
- Document Sentinel configuration format in detail, including required fields, defaults, sink routing, and environment variable expansion behavior.
- Document notifier sinks, including `stdout/debug` behavior and webhook delivery semantics.
- Add practical run/test/troubleshooting guides for local validation.
- Rewrite the root `README.md` in a conventional project format without emojis, with concise quick-start instructions and links into the docs site.

## Capabilities

### New Capabilities
- `docsify-operator-guides`: Operator documentation in a Docsify-compatible `docs/` tree covering configuration, sinks, runtime behavior, and troubleshooting.

### Modified Capabilities
- `sentinel-cli-config`: Clarify and document CLI/config usage contract, including config file examples and env-var interpolation expectations.
- `notification-delivery-pipeline`: Document sink behavior and delivery semantics so operators can reason about stdout/debug and webhook outputs.

## Impact

- Affected code/assets: `docs/` content and structure, root `README.md`, and any examples referenced by docs.
- No runtime API or protocol changes are required.
- Improves onboarding and operability while reducing ambiguity around configuration and sink behavior.
