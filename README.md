# Seer

**Seer** is a self-hosted tool to detect market opportunities for indie developers.

> "The Seer sees what others miss."

## Philosophy

- You control your data
- You choose where it runs

## Quick Start

### Using Docker

```bash
docker-compose up
```

### Using Binary

```bash
# Download the latest release
./seer
```

Visit `http://localhost:8080` to access the dashboard.

### Downloads

Download the latest binary for your platform:

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | x64 | [seer-linux-amd64](https://github.com/mx-seer/seer/releases/latest/download/seer-linux-amd64) |
| Linux | ARM64 | [seer-linux-arm64](https://github.com/mx-seer/seer/releases/latest/download/seer-linux-arm64) |
| macOS | Intel | [seer-darwin-amd64](https://github.com/mx-seer/seer/releases/latest/download/seer-darwin-amd64) |
| macOS | Apple Silicon | [seer-darwin-arm64](https://github.com/mx-seer/seer/releases/latest/download/seer-darwin-arm64) |
| Windows | x64 | [seer-windows-amd64.exe](https://github.com/mx-seer/seer/releases/latest/download/seer-windows-amd64.exe) |

Or visit the [Releases page](https://github.com/mx-seer/seer/releases) for all versions.

## Configuration

Copy the example configuration:

```bash
cp config.example.yaml config.yaml
```

### Configuration Options

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "./data/seer.db"

# Sources configuration
sources:
  fetch_interval: 60  # Interval in minutes between source fetches (default: 60)
```

## Features

- **Dashboard** - View and manage detected opportunities
- **Manual Refetch** - Trigger source fetching on-demand via dashboard
- **Configurable Fetch Interval** - Set how often sources are checked (default: every hour)
- **SQLite Database** - Lightweight, self-contained data storage with WAL mode

## Development

### Prerequisites

- Go 1.23+
- Bun (for frontend)

### Commands

```bash
# Run in development mode (frontend + backend)
make dev

# Run Go server only
make dev-server

# Build binary
make build

# Run tests
make test

# Clean build artifacts
make clean
```

## Sources

Seer monitors multiple sources for opportunities:

| Source | Description |
|--------|-------------|
| Hacker News | Algolia API |
| GitHub | Trending/Issues |
| npm | New packages |
| DEV.to | Articles |

## Tech Stack

- **Backend**: Go 1.23+ with chi router
- **Frontend**: SvelteKit with Svelte 5, Tailwind CSS v4, daisyUI 5.x
- **Database**: SQLite with WAL mode
- **Build**: Makefile with embedded frontend

## License

See [LICENSE](LICENSE) file.

## Support

- Email: seer@mendex.io
- Issues: [GitHub Issues](https://github.com/mx-seer/seer-pro/issues)
