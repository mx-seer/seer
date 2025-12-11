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
