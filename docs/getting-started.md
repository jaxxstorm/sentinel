# Getting Started

Sentinel supports three run paths:

1. GitHub Release binary (recommended for operators)
2. Docker / Docker Compose
3. Source run with `go run` (development)

## Prerequisites

- A Tailscale account/tailnet
- Local access to persist Sentinel tsnet state
- One of:
  - installed Sentinel binary
  - Docker
  - Go 1.23+ (development path only)

## Choose a Run Path

| Path | Best for | Entry point |
| --- | --- | --- |
| GitHub Release binary | Operator installs on host/VM | `sentinel ...` |
| Docker / Compose | Containerized deployment | `docker run ...` or `docker compose ...` |
| Source run | Local development/debugging | `go run ./cmd/sentinel ...` |

## Path A: Install from GitHub Releases (Recommended)

Choose the release tag and asset pattern for your platform:

- Linux amd64: `sentinel_*_linux_amd64.tar.gz`
- Linux arm64: `sentinel_*_linux_arm64.tar.gz`
- macOS amd64: `sentinel_*_darwin_amd64.tar.gz`
- macOS arm64: `sentinel_*_darwin_arm64.tar.gz`
- Windows amd64: `sentinel_*_windows_amd64.zip`

Example install (Linux amd64):

```bash
VERSION=v0.1.0

gh release download "$VERSION" \
  --repo jaxxstorm/sentinel \
  --pattern 'sentinel_*_linux_amd64.tar.gz' \
  --pattern 'checksums.txt'

tar -xzf sentinel_*_linux_amd64.tar.gz sentinel
install -m 0755 sentinel /usr/local/bin/sentinel
sentinel version
```

If you do not use GitHub CLI, download matching assets manually from:
`https://github.com/jaxxstorm/sentinel/releases`

Then verify configuration and run:

```bash
sentinel validate-config --config ./config.example.yaml
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
sentinel test-notify --config ./config.example.yaml
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
sentinel run --config ./config.example.yaml --log-level info
```

## Path B: Docker / Compose

For local container workflows with source builds:

```bash
cp .env.example .env
```

Set `SENTINEL_TAILSCALE_AUTH_KEY` in `.env`, then run:

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml up --build
```

For GHCR image usage and Railway template import guidance, see:

- [Docker Compose and Railway](docker-compose.md)
- [Docker Image](docker-image.md)

## Path C: Source Run (Development)

This path requires Go 1.23+ and is intended for local development:

```bash
go run ./cmd/sentinel version
```

```bash
go run ./cmd/sentinel validate-config --config ./config.example.yaml
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
