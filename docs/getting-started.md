# Getting Started

## Prerequisites

- Go 1.23+
- A Tailscale account/tailnet
- Local access to run Sentinel with tsnet state storage

## Quick Start

```bash
go run ./cmd/sentinel version
```

```bash
go run ./cmd/sentinel validate-config --config ./config.example.yaml
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
go run ./cmd/sentinel test-notify --config ./config.example.yaml
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
go run ./cmd/sentinel run --config ./config.example.yaml --log-level info
```

## Local Auth Modes

Sentinel supports three onboarding modes in `tsnet.login_mode`:

- `auto`: use auth key if available, otherwise interactive login
- `auth_key`: require auth key
- `interactive`: always use login URL flow

Auth key sources are checked in this order:

1. `--tailscale-auth-key`
2. `SENTINEL_TAILSCALE_AUTH_KEY`
3. `tsnet.auth_key` from config

## Local Docs Preview (Docsify)

If you have Docsify CLI installed:

```bash
docsify serve docs
```

Then open `http://localhost:3000`.
