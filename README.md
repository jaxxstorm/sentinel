# Sentinel

Sentinel is a tsnet-embedded Tailscale observer. It tracks tailnet netmap changes, detects meaningful diffs, and sends notifications through configurable sinks.

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/deploy/OyPzYt?referralCode=ftkvtR&utm_medium=integration&utm_source=template&utm_campaign=generic)

## Features

- Realtime observation via Tailscale IPNBus (`source.mode: realtime`)
- Optional polling mode (`source.mode: poll`)
- Presence event detection (`peer.online`, `peer.offline`)
- Route-based notifier pipeline with multiple sinks
- Always-on local JSON sink (`stdout-debug`) for visibility
- Webhook delivery with retries and structured success/failure logging
- Structured logging with stable `log_source` attribution (`sentinel`, `tailscale`, `sink`)

## Installation Paths

- GitHub Release binary (recommended for operators)
- Docker image / Docker Compose
- Source run with `go run` (development)

See [`docs/getting-started.md`](docs/getting-started.md) for complete setup details.

## Quick Start (GitHub Release Binary)

Download and install from GitHub Releases (Linux amd64 example):

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

Then run Sentinel:

```bash
sentinel validate-config --config ./config.example.yaml
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
sentinel test-notify --config ./config.example.yaml
```

```bash
REQUESTBIN_WEBHOOK_URL="https://your-endpoint" \
sentinel run --config ./config.example.yaml
```

## Quick Start (Docker)

```bash
cp .env.example .env
```

Set `SENTINEL_TAILSCALE_AUTH_KEY` in `.env`, then run:

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml up --build
```

For GHCR image usage, Railway import, and environment matrix details, see [`docs/docker-compose.md`](docs/docker-compose.md) and [`docs/docker-image.md`](docs/docker-image.md).

## Configuration

- Example config: [`config.example.yaml`](config.example.yaml)
- Supports YAML/JSON with `SENTINEL_` environment overrides
- Supports `${VAR_NAME}` interpolation in sink URLs

## Commands

- `run`
- `status`
- `diff`
- `dump-netmap`
- `test-notify`
- `validate-config`

Use `sentinel --help` for full command and flag details.

## Documentation

Operator docs live under [`docs/`](docs/README.md) and are structured for Docsify.

- [Getting Started](docs/getting-started.md)
- [Configuration Reference](docs/configuration.md)
- [Docker Compose and Railway](docs/docker-compose.md)
- [Docker Image](docs/docker-image.md)
- [Release Artifacts](docs/release-artifacts.md)
- [Sinks and Routing](docs/sinks-and-routing.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Command Reference](docs/commands.md)

To preview with Docsify:

```bash
docsify serve docs
```

## Development

Run tests:

```bash
go test ./...
```

Run from source with `go run`:

```bash
go run ./cmd/sentinel run --config ./config.example.yaml --log-level debug
```
