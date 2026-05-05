# Changelog

All notable changes to ProjectArgus will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased] — targeting v0.2.0

### Added
<!-- items from v0.2.0 milestone -->

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
