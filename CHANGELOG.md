# Changelog

All notable changes to ProjectArgus will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- Atlas service (M1–M3): Go module with Chi server, hello-world page, SSE endpoint,
  NATS/JetStream subscribers, Tailscale sources, service probe matrix, and hosts page
- `.dockerignore` to minimise Docker build context

### Changed

- CI lint job consolidates typecheck steps; unified required-checks workflow added
- Python base image bumped to `3.14.0-slim`

### Fixed

- Atlas SSE heartbeat data race in tests
- Atlas M2 review-wave findings (healthz plaintext, hermes port, README line wrap)
- CI: valid job IDs in `_required.yml` (no slashes allowed in keys)
- Dependabot: removed invalid Docker `package-ecosystem` entry

## [0.1.0] - 2026-04-23

### Added

- Prometheus scrape stack with 15s global interval
- Loki log aggregation with 30-day retention
- Promtail log shipping from container stdout and host log files
- Grafana dashboards: agent-health, nats-events, task-throughput
- Custom Python exporter (`exporter/exporter.py`) scraping ProjectAgamemnon and
  NATS HTTP APIs, exposing metrics on port 9101
- Prometheus alerting rules in `rules/agent-alerts.yml` for agent health
- Grafana Alertmanager contact point provisioning
- Docker Compose orchestration with pinned image versions
- `justfile` task runner with targets for start, stop, reload, scrape, and dashboard import
- CI workflow (yamllint + Python lint) on GitHub Actions
- CLAUDE.md project conventions and development guidelines
- CONTRIBUTING.md, SECURITY.md, LICENSE at repository root
- `.gitignore` covering `.pixi/`, IDE files, OS files, and secrets
- `pixi.toml` with locked dependencies (`just`, `jq`)

[Unreleased]: https://github.com/HomericIntelligence/ProjectArgus/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/HomericIntelligence/ProjectArgus/releases/tag/v0.1.0
