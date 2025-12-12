<p align="center">
  <img src="web/static/favicon.svg" alt="Seer Logo" width="80" height="80">
</p>

<h1 align="center">Seer</h1>

<p align="center">
  <strong>See what others miss</strong><br>
  Market opportunity detection for indie developers
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT_+_Commons_Clause-blue.svg" alt="License">
  <img src="https://img.shields.io/github/v/release/mx-seer/seer?color=purple" alt="Release">
  <img src="https://img.shields.io/github/stars/mx-seer/seer?style=flat&color=yellow" alt="Stars">
  <img src="https://img.shields.io/github/issues/mx-seer/seer" alt="Issues">
</p>

---

## What is Seer?

Seer monitors Hacker News, GitHub, npm, and DEV.to to surface market opportunities—pain points, feature requests, trending projects, and gaps you can fill. Self-hosted, privacy-first, no tracking.

## Features

- **Multi-source monitoring** — HN, GitHub, npm, DEV.to with optimized queries
- **Opportunity scoring** — AI-powered relevance scoring (0-100)
- **Real-time dashboard** — Filter, search, and track opportunities
- **Self-hosted** — Your data stays on your machine
- **Configurable** — Fetch intervals, source weights, score thresholds

## Quick Start

### Docker (recommended)

```bash
docker run -d -p 8080:8080 -v seer-data:/app/data ghcr.io/mx-seer/seer:latest
```

### Binary

Download from [Releases](https://github.com/mx-seer/seer/releases/latest), then:

```bash
./seer
```

Visit `http://localhost:8080`

## Configuration

```bash
cp config.example.yaml config.yaml
```

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "./data/seer.db"

sources:
  fetch_interval: 60  # minutes
```

## Sources

| Source | What it finds |
|--------|--------------|
| Hacker News | Pain points, wishes, alternatives requests |
| GitHub | Trending repos, help wanted, self-hosted projects |
| npm | New packages, dev tools, SDKs |
| DEV.to | Show projects, discussions, tutorials |

## Tech Stack

**Backend**
![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-003B57?style=flat&logo=sqlite&logoColor=white)

**Frontend**
![SvelteKit](https://img.shields.io/badge/SvelteKit-FF3E00?style=flat&logo=svelte&logoColor=white)
![Svelte](https://img.shields.io/badge/Svelte_5-FF3E00?style=flat&logo=svelte&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_v4-06B6D4?style=flat&logo=tailwindcss&logoColor=white)
![daisyUI](https://img.shields.io/badge/daisyUI-5A0EF8?style=flat&logo=daisyui&logoColor=white)

**Infrastructure**
![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white)

## Support the Project 💜

Seer is **100% free and open source**. No Pro tiers, no paywalls.

If Seer helps you discover opportunities, consider supporting development:

[![GitHub Sponsors](https://img.shields.io/badge/GitHub_Sponsors-EA4AAA?style=flat&logo=github-sponsors&logoColor=white)](https://github.com/sponsors/mendexio)
[![Ko-fi](https://img.shields.io/badge/Ko--fi-FF5E5B?style=flat&logo=ko-fi&logoColor=white)](https://ko-fi.com/mendexio)

## Development

```bash
make dev          # Run frontend + backend
make build        # Build binary
make test         # Run tests
```

## Contributing

Found a bug or have a suggestion? [Open an issue](https://github.com/mx-seer/seer/issues).

## License

[MIT with Commons Clause](LICENSE) — free to use, modify, and build upon. Reselling the unmodified software is not permitted.

---

<p align="center">
  Built with 💜 by <a href="https://mendex.io">Mendex</a>
</p>
