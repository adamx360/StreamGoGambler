# StreamGoGambler

A simple tool that automatically plays slots and joins heists on Twitch for you. **No coding or technical skills required!**

---

## What does this bot do?

StreamGoGambler is like a "helper" for your Twitch chat. Once you set it up, it will:

- **Play Slots Automatically:** It watches your points and plays slots whenever you have enough.
- **Join !heist, !boss, !ffa etc. :** It automatically enters events for you.
- **Track Your Points:** It keeps an eye on your balance so you don't have to.
- **Stay Running in the Background:** You can minimize it to your system tray (near the clock) and let it do its magic.

---

## Quick Start Guide

### 1. Download & Install

#### Windows (Easiest)
1. **Download:** Go to the [Latest Releases](https://github.com/adamx360/streamgogambler/releases) and download `streamgogambler-windows-amd64.exe`.
2. **Setup Folder:** Create a new folder (e.g., `C:\TwitchBot`) and move the downloaded file there.
3. **Run:** Double-click the file. 
   - *Note: If a blue "Windows protected your PC" box appears, click **"More info"** and then **"Run anyway"**.*

#### Mac
1. **Download:** Go to the [Latest Releases](https://github.com/adamx360/streamgogambler/releases) and download the version for your Mac (`arm64` for newer M1/M2/M3 Macs, `amd64` for older Intel Macs).
2. **First Run:** Right-click the file and select **"Open"**. When the warning appears, click **"Open"** again.

---

### 2. Getting Your Login Token (OAuth)

The bot needs a "key" to talk in chat on your behalf. This is called an **OAuth Token**.

1. Go to [TwitchTokenGenerator.com](https://twitchtokengenerator.com/)
2. Click the big **"Bot Chat Token"** button.
3. Log in with the Twitch account you want the bot to use.
4. Click **"Authorize"**.
5. **Copy the "Access Token"** (it looks like a long string of random letters and numbers).

   **CRITICAL:** Treat this token like a password. Never show it to anyone!

---

### 3. First-Time Setup

When you open the bot for the first time, a window will pop up. Fill it out like this:

- **Twitch Username:** Your Twitch name.
- **OAuth Token:** Paste the token you just got.
- **Channel:** The name of the streamer's channel where you want to play.
- **Command Prefix:** Usually `!`
- **Boss Bot Name:** The name of the bot that manages points in that channel (usually `StreamElements` or `Nightbot`).

Click **"Save & Start Bot"** - you're done!

---

## Using the Bot

- **Main Window:** Shows your points, how long the bot has been running, and a log of what it's doing.
- **Closing the Window:** If you click the **X**, the bot doesn't stop! It hides in your **System Tray** (near the clock on Windows). 
- **Right-Click the Icon:** Right-click the small bot icon near your clock to **Show** the window again or **Quit** the bot completely.

### How to turn on "Auto-Slots"
By default, the bot won't play slots until you tell it to.
1. Go to the Twitch chat where the bot is.
2. Type: `!autoslots on`
3. The bot will now start playing automatically!

---

## Common Commands

Type these in the Twitch chat:

| Command          | What it does                                        |
|:-----------------|:----------------------------------------------------|
| `!status`        | Checks if the bot is working and shows your points. |
| `!autoslots on`  | Turns on automatic slot playing.                    |
| `!autoslots off` | Stops automatic slot playing.                       |
| `!ustaw 500`     | Changes your heist bet to 500 points.               |
| `!help`          | Shows a list of all things the bot can do.          |

---

## Troubleshooting

### "Windows protected your PC"
This is normal! Click **"More info"** then **"Run anyway"**. We aren't a big company, so Windows doesn't recognize us yet.

### The bot isn't playing slots!
- Make sure you typed `!autoslots on` in chat.
- Check if you have enough points (usually you need at least 2000).
- Make sure the "Boss Bot Name" in settings matches the channel's bot exactly.

### How do I change my settings later?
Look in the folder where you put the bot. You'll see a file named `.env`. 
- **To start over:** Delete the `.env` file and run the bot again. The setup window will reappear.
- **To edit:** Open `.env` with Notepad and change the values.

---

## How to Uninstall
1. Right-click the bot icon near your clock and choose **Quit**.
2. Delete the folder you created for the bot.
3. That's it! No messy "Add/Remove Programs" needed.

---

## For Developers

<details>
<summary>Click to expand technical documentation</summary>

### Badges

[![CI](https://github.com/adamx360/streamgogambler/actions/workflows/ci.yml/badge.svg)](https://github.com/adamx360/streamgogambler/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Mutation Testing](https://img.shields.io/badge/mutation%20testing-85%25-green.svg)](https://github.com/go-gremlins/gremlins)

### Technical Features

- **Graphical User Interface** - Optional Fyne-based GUI with real-time stats display
- **System tray support** - Minimize to system tray, runs in background
- **Single instance lock** - Prevents multiple copies from running simultaneously
- **First-time setup wizard** - GUI dialog for easy initial configuration
- **Thread-safe state management** - Mutex-protected shared state for concurrent access
- **Smart command gating** - Paid commands only execute when balance covers the cost
- **Trusted sender validation** - Parses messages only from configured boss bot
- **Rate limiting** - Token-bucket rate limiting with configurable burst and refill
- **Health monitoring** - HTTP endpoint for monitoring bot status
- **Graceful shutdown** - Clean shutdown with OS signal handling
- **Automatic reconnection** - Exponential backoff retry on connection failures
- **Structured logging** - Uses Go's standard `log/slog` for better observability
- **Full context support** - Context propagation throughout the application for better tracing

### Architecture

This project follows **Domain-Driven Design (DDD)** with **Hexagonal Architecture** (Ports & Adapters):

```
streamgogambler/
├── cmd/streamgogambler/          # Entry point, dependency wiring
├── internal/
│   ├── domain/             # Core business logic (no external deps)
│   │   ├── parsing/        # Message parsing (bombs, slots, points)
│   │   ├── wallet/         # Currency balance entity
│   │   └── gambling/       # Heist rules and validation
│   ├── application/        # Use cases, orchestration
│   ├── ports/              # Interfaces (contracts)
│   └── adapters/           # Infrastructure implementations
│       ├── twitch/         # IRC client wrapper
│       ├── config/         # .env loading & persistence
│       ├── gui/            # Fyne-based graphical interface
│       ├── healthcheck/    # Health endpoint
│       ├── logging/        # Leveled logging (using slog)
│       └── storage/        # Trusted users persistence
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/adamx360/streamgogambler.git
cd streamgogambler

# Copy and configure environment
cp .env.example .env
# Edit .env with your Twitch credentials

# Build and run
make build
./streamgogambler
```

Or using Go directly:
```bash
go build ./cmd/streamgogambler
./streamgogambler
```

### Configuration Reference

#### Required Variables

| Variable          | Description                           |
|-------------------|---------------------------------------|
| `TWITCH_USERNAME` | Bot account username                  |
| `TWITCH_OAUTH`    | OAuth token (without `oauth:` prefix) |
| `TWITCH_CHANNEL`  | Channel to join                       |
| `COMMAND_PREFIX`  | Command prefix (e.g., `!`)            |
| `STATUS_COMMAND`  | Status command name                   |
| `CONNECT_MESSAGE` | Message sent on first connect         |
| `BOSS_BOT_NAME`   | Trusted bot for parsing messages      |

#### Optional Variables

| Variable              | Default | Description                                        |
|-----------------------|---------|----------------------------------------------------|
| `HEIST_AMOUNT`        | 1000    | Default heist amount                               |
| `SLOTS_COST`          | 2000    | Cost per !slots command                            |
| `ARENA_COST`          | 1000    | Cost per !ffa command                              |
| `AUTO_SLOTS_ENABLED`  | false   | Is autoslots enable on startup                     |
| `AUTO_SLOTS_INTERVAL` | 15      | Autoslots interval in minutes                      |
| `BAND_ON_PERMA`       | false   | Send message on permanent bans                     |
| `BAND_MESSAGE`        | BAND    | Ban response message                               |
| `POINTS_AS_DELTA`     | true    | Treat points as delta vs absolute                  |
| `SAY_BUCKET_SIZE`     | 20      | Token bucket size for rate limiting                |
| `SAY_REFILL_MS`       | 150     | Token refill interval (ms)                         |
| `GREET_ON_RECONNECT`  | false   | Send greeting after reconnects                     |
| `LOG_LEVEL`           | info    | Log verbosity: debug, info, warn, error            |
| `HEALTH_PORT`         | 0       | Health endpoint port (0 = disabled)                |
| `GUI_ENABLED`         | true    | Enable graphical interface (false = headless mode) |
| `MAX_LOGS_LINES`      | 500     | # Maxiumum number of log lines in gui              |

#### Configuration Precedence

The `.env` file is loaded from:
1. Path specified in `ENV_PATH` environment variable
2. Directory containing the executable
3. Current working directory (fallback)

### Health Endpoint

Enable monitoring by setting `HEALTH_PORT`:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "uptime": "1h30m15s",
  "uptime_seconds": 5415,
  "bombs": 12500,
  "messages_sent": 42,
  "messages_received": 1337,
  "reconnect_count": 0,
  "channel": "yourchannel",
  "username": "yourbotname"
}
```

### Development

#### Prerequisites

- Go 1.25+
- golangci-lint (for linting)
- pre-commit (optional, for git hooks)

#### Setup

```bash
# Install development tools
make install-tools

# Install pre-commit hooks (optional)
pip install pre-commit
make pre-commit-install
```

#### Commands

```bash
# Build
make build          # Build binary with version info
make build-dev      # Build without version info (faster)
make build-all      # Build for all platforms (Windows, Linux, macOS)
make run            # Build and run the bot

# Testing
make test           # Run all tests
make test-short     # Run tests in short mode
make test-race      # Run tests with race detector
make test-cover     # Generate coverage report
make fuzz           # Run fuzz tests (10s each)
make fuzz-long      # Run extended fuzz tests (60s each)

# Code Quality
make lint           # Run golangci-lint
make lint-fix       # Run linter with auto-fix
make fmt            # Format code with gofmt
make vet            # Run go vet

# Mutation Testing
make mutate         # Run mutation testing (quick)
make mutate-dry     # Analyze mutations without running tests
make mutate-full    # Run full mutation testing with all mutant types

# Docker
make docker-build   # Build Docker image
make docker-run     # Run bot in Docker container
make docker-stop    # Stop Docker container
make docker-logs    # Show Docker container logs

# Utilities
make deps           # Verify and tidy dependencies
make check          # Run all checks (lint, test, build)
make ci             # Run full CI pipeline locally
make clean          # Remove build artifacts
make version        # Show version info
make help           # Show all commands
```

### Releases

Releases are automated via GitHub Actions:

1. **Update `CHANGELOG.md`** with your new version (format: `## [x.x.x]`)
2. **Merge to master** - When a PR with CHANGELOG.md changes is merged:

- The `auto-tag` workflow extracts the version and creates a Git tag (e.g., `v1.2.0`)
- The `release` workflow triggers automatically on the new tag
- Binaries are built for Linux, Windows, and macOS
- A GitHub Release is created with all artifacts

### Parsing Language

The bot recognizes outcomes based on English phrases from the boss bot:

- **Slots**: `you lost`, `jackpot`, `super jackpot`, `even a small win is a win..`
- **Points**: `<User> (<number>)` format, e.g., `UserX (2 000)`
- **Bombs**: `<user> bombs: <number>`

To customize parsing for different languages, modify the parsers in `internal/domain/parsing/`.

### Security

- **Trusted sender validation** - Only processes messages from the configured boss bot
- **Input validation** - UTF-8 validation and length limits prevent injection attacks
- **No hardcoded secrets** - All credentials loaded from environment variables
- **Automated scanning** - CI includes gosec and govulncheck security scans
- **Dependency monitoring** - Dependabot configured for automated security updates

</details>

---

## License

[MIT](LICENSE) - Copyright (c) 2026 Adamx360

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
