# Changelog

All notable changes to StreamGoGambler will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-31

### Added

- **Graphical User Interface** - Optional Fyne-based GUI with real-time statistics display
  - Connection status, balance, messages sent/received, uptime display
  - Activity log showing live bot output
  - Command input field to execute commands directly from GUI
  - Auto slots toggle checkbox
- **System tray support** - Minimize to system tray instead of closing
  - Right-click menu with Show, Hide, and Quit options
  - Bot continues running in background when window is closed
- **First-time setup wizard** - GUI dialog for initial configuration
  - Prompts for required settings when .env is missing
  - Saves configuration and starts bot automatically
- **Single instance lock** - Prevents multiple copies from running simultaneously
  - Windows: Uses named mutex for process-level locking
  - Shows error dialog if another instance is already running
- **Application icon** - Custom icon for GUI window and Windows executable
  - Embedded icon in GUI window title bar
  - Windows binary shows icon in Explorer and taskbar
- **Automatic release tagging** - GitHub Actions workflow that automatically creates Git tags
- **X11 development libraries** - Added `libx11-dev` installation step for Linux GUI builds
- **Migrated logging to slog** - Replaced custom logging implementation with Go's standard `log/slog` package
  - Provides structured logging capabilities
  - Better integration with Go ecosystem tooling
- **Context.Context in logging** - All logging methods now accept `context.Context` for better tracing
- `GUI_ENABLED` environment variable to toggle GUI on/off (default: true)
- Headless mode for server deployments (GUI_ENABLED=false)
- Thread-safe bombs currency management with mutex protection
- Integration with boss bot for automated responses
- Command system with trusted user access control:
  - `!status` - Show bot status and balance (trusted)
  - `!ustaw <amount>` - Set default heist amount (trusted)
  - `!jakiheist` - Show current heist amount (trusted)
  - `!autoslots on/off` - Enable/disable auto slots (trusted)
  - `!slotsoff <time/duration>` - Schedule auto slots turn off (trusted)
  - `!help` - List available commands (trusted)
  - `!trust <user>` - Add user to trusted list (owner only)
  - `!untrust <user>` - Remove user from trusted list (owner only)
  - `!trustlist` - Show trusted users (owner only)
- Trusted users whitelist system (owner + whitelisted users can use commands)
- Auto slots feature with 15-minute cooldown (disabled by default)
- Token-bucket rate limiting for outgoing messages
- Per-user command rate limiting (2s interval, owner exempt)
- Auto-response system for boss bot triggers
- Configurable auto-responses via environment variables
- Health check HTTP endpoint (`/health`) for monitoring
- Graceful shutdown with OS signal handling (SIGINT/SIGTERM)
- Exponential backoff for connection retries (1s to 5min)
- Jittered background tasks to prevent thundering herd
- Atomic `.env` file updates with trailing newline preservation
- Configurable logging levels (debug, info, warn, error)
- Comprehensive parser functions for bombs, slots, and points
- UTF-8 validation and length limits in parsers
- **README documentation improvements** - Expanded Makefile and required variables documentation

### Security

- Trusted sender validation (only parses messages from configured boss bot)
- Paid command gating (prevents sending when balance is insufficient)
- `.env` file excluded from version control
- No hardcoded credentials

### Testing

- 85%+ mutation testing efficacy (gremlins)
- Domain packages: 100% mutator coverage
- Storage adapter: 100% mutator coverage
- Comprehensive boundary tests for all edge cases
- Unit tests for all parser functions
- Integration tests with fake IRC client
- Fuzz tests for parser robustness
- Concurrent access tests for thread safety
- Timing-resilient tests using polling instead of fixed delays

### Documentation

- Comprehensive README with configuration guide
- Inline code documentation and Godoc comments
- Configuration precedence documentation
- Rate limiting tuning guide
- Health endpoint usage examples

[1.0.0]: https://github.com/REPLACEME/streamgogambler/releases/tag/v1.0.0
