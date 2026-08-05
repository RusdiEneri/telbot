# Telbot

Go-based tool for managing Telkomsel accounts via **Telegram Bot**, **Terminal CLI**, or **MCP Server** (for AI agents).

## Features

- 🔑 **SMS OTP login** with session caching
- 👥 **Multi-account support** (Switch & manage multiple Telkomsel numbers seamlessly)
- 📊 **Profile, balance & quota checking**
- 📦 **Browse recommended packages**
- 🛍️ **Purchase packages** (Pulsa / QRIS)
- ⏰ **Auto-buy monitor** (Auto-purchase when quota is depleted or below custom MB threshold)
- 🔄 **Auto re-login via OTP webhook** (Session renewal without manual intervention)
- 🤖 **MCP server** for AI agent integration
- 🐳 **Docker & Cloud ready** (Easy deployment on Fly.io, Railway, VPS with persistent storage support)

## Quick Start

```bash
git clone https://github.com/0xtbug/telbot
cd telbot
cp .env.example .env   # fill in your tokens
go mod tidy
```

## Modes

| Flag | Description |
|------|-------------|
| `--bot` | Telegram bot (requires `.env` config) |
| `--cli` | Interactive terminal UI (Bubbletea) |
| `--mcp` | MCP server for AI agents (stdio) |

```bash
go run . --bot       # Telegram bot
go run . --cli       # Terminal UI
go run . --mcp       # MCP server
```

## Installation (Pre-built Binaries)

If you don't want to install Go or build it yourself, download the pre-compiled executables from the **[Releases](https://github.com/0xtbug/telbot/releases)** page.

**Linux / macOS:**
```bash
chmod +x telbot-linux-amd64
sudo mv telbot-linux-amd64 /usr/local/bin/telbot
# now you can run `telbot` from anywhere
```

**Windows:**
1. Download the `.exe` from Releases.
2. Rename it to `telbot.exe` and move to a permanent folder (e.g. `C:\telbot\`).
3. Add that folder to your `PATH` via *System → Environment Variables*.

## 🐳 Docker & Cloud Deployment

### Using Docker

```bash
# Build
docker build -t telbot .

# Run with persistent data volume
docker run -d \
  --name telbot \
  --restart unless-stopped \
  -e TELKOMSEL_BOT_TOKEN="your_bot_token" \
  -e TELEGRAM_ADMIN_ID="your_telegram_id" \
  -e TELBOT_DATA_DIR="/data" \
  -p 8080:8080 \
  -v telbot_data:/data \
  telbot
```

> 💡 **Tip:** Put your `.env` inside the volume (`/data/.env`) so it persists across container restarts. The entrypoint will auto-load it.

### Deploy to Fly.io (Recommended)

Fly.io supports **persistent volumes** out of the box — perfect for storing `sessions.json`.

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Login
flyctl auth login

# Create app + volume (one-time setup)
flyctl apps create telbot
flyctl volumes create telbot_data --region sin --size 1

# Deploy
flyctl deploy
```

Your `fly.toml` should look like this:

```toml
app = "telbot"
primary_region = "sin"

[build]

[env]
  TELBOT_DATA_DIR = "/data"

[mounts]
  source = "telbot_data"
  destination = "/data"

[[services]]
  internal_port = 8080
  protocol = "tcp"
```

### Deploy to Railway

1. Connect your GitHub repo to [Railway](https://railway.app/).
2. Add environment variables (`TELKOMSEL_BOT_TOKEN`, `TELEGRAM_ADMIN_ID`).
3. Add a **Volume** and mount it to `/data`.
4. Set `TELBOT_DATA_DIR=/data` in environment variables.
5. Deploy.

### Deploy to VPS (DigitalOcean, Vultr, Hostinger, etc.)

```bash
# On your VPS
git clone https://github.com/0xtbug/telbot
cd telbot
cp .env.example .env
nano .env  # fill in tokens
go build -o telbot .
sudo mv telbot /usr/local/bin/

# Use systemd to keep it running 24/7 (see docs/systemd.md)
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `TELKOMSEL_BOT_TOKEN` | Bot mode | Telegram bot token from BotFather |
| `TELEGRAM_ADMIN_ID` | Bot mode | Your Telegram user ID |
| `TELBOT_DATA_DIR` | Optional | Custom directory for storing `sessions.json` and `.env` (e.g. `/data`) |
| `OTP_WEBHOOK_PORT` | Optional | Port for OTP webhook listener (e.g. `8080`) |
| `OTP_WEBHOOK_SECRET` | Optional | Shared secret for webhook authentication |

### `.env` lookup order

Telbot will try to load `.env` from these locations (in order):

1. `$TELBOT_DATA_DIR/.env` (if `TELBOT_DATA_DIR` is set)
2. Current working directory
3. The directory where the binary lives
4. `~/.config/telbot/.env` (Linux/macOS) or `%APPDATA%\telbot\.env` (Windows)

## 🤖 AI Agent Integration (OpenClaw)

This repository includes official support for [OpenClaw](https://github.com/openclaw/openclaw), an open-source AI agent framework. Prompt your AI with:

```text
Please load and use this telbot Skill:
https://raw.githubusercontent.com/0xtbug/telbot/main/telbot-skills/SKILL.md
```

## Documentation

See the [docs/](docs/) folder for detailed guides:

- [Telegram Bot](docs/telegram-bot.md)
- [CLI Mode](docs/cli.md)
- [MCP Server](docs/mcp-server.md)
- [Auto Re-login (OTP Webhook)](docs/auto-relogin.md)
- [Systemd Service (Linux)](docs/systemd.md)
- [OpenWrt](docs/openwrt.md)
- [OpenClaw Skill](telbot-skills/SKILL.md)

## Disclaimer

This project is an unofficial tool that interacts with MyTelkomsel web services for personal automation purposes.

It does not modify, bypass, or exploit any security mechanisms of the MyTelkomsel platform. All requests are performed using standard HTTP interactions similar to those made by the official web interface.

This project is intended for educational and personal use only. Use at your own risk.
