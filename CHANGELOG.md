# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.4] - 2025-12-11

### Fixed

- Stats card now respects min_score and source filters

## [0.1.3] - 2025-12-11

### Fixed

- Fixed GitHub Issues link in README

## [0.1.2] - 2025-12-11

### Fixed

- Simplified download links in README

## [0.1.1] - 2025-12-11

### Added

- MIT License
- Download links in README

## [0.1.0] - 2025-12-11

### Added

- Initial project structure
- Configuration loader (YAML)
- SQLite database with migrations and WAL mode
- Chi router with health check endpoint
- Docker support
- SvelteKit frontend with Svelte 5, Tailwind CSS v4, daisyUI 5.x
- Dashboard to view and manage opportunities
- Manual refetch button to trigger source fetching on-demand
- Configurable fetch interval via config file (default: 60 minutes)
- Source manager with scheduled fetching (cron-based)
- Sources: Hacker News, GitHub, npm, DEV.to
- Multi-platform build support (Linux, macOS, Windows)
