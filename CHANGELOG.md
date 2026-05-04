# Changelog

All notable changes to ProjectArgus will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.2.0] - 2026-05-04

### Changed

- **Exporter Dockerfile**: multi-stage build, non-root `exporter` user, `HEALTHCHECK`
  instruction — addresses `[MAJOR] §12` finding (#148)
- **Metric naming**: renamed `homeric_exporter_scrape_timestamp` →
  `homeric_exporter_scrape_timestamp_seconds` to follow Prometheus naming conventions
  — addresses `[MINOR] §14` finding (#156)
- **Alert rule**: updated `ExporterScrapeStale` to reference renamed timestamp metric
- **pixi.toml**: added `test` task (`python -m pytest tests/ -v`), broadened platforms
  to include `osx-arm64`, `osx-64`, `win-64` — addresses `[MINOR] §13` (#153) and
  `[MINOR] §12` (#151)
- **justfile**: removed hardcoded `admin:admin` from `import-dashboards` (now reads
  `GF_ADMIN_PASSWORD` from `.env`); added `.env` existence check to `start` and
  `import-dashboards`; replaced `docker exec`-based `test-scrape` with
  `docker compose exec` — addresses `[MAJOR] §13` (#152) and §13 (#187)
- **CLAUDE.md**: updated stale image version table, added missing `NATS_LOG_DIR`
  environment variable, added Mermaid architecture diagram, metric naming section,
  and AI agent collaboration notes — addresses `[MAJOR] §11` (#144) and
  `[MINOR] §11` (#147)
- **SECURITY.md**: replaced GitHub no-reply address with real contact email
  — addresses `[MINOR] §15` (#159)
- **LICENSE**: updated copyright year from 2025 → 2026
  — addresses `[MINOR] §15` (#158)

### Added

- **`# HELP` lines** in exporter `/metrics` output for every metric
  — addresses `[MINOR] §14` (#155)
- **`AGENTS.md`**: multi-agent coordination protocol and permitted-change matrix
  — addresses `[MINOR] §11` (#145)
- **`CODEOWNERS`**: maps all files to `@mvillmow` with CI/security escalations
  — addresses `[MINOR] §10` (#142)
- **`.github/PULL_REQUEST_TEMPLATE.md`**: structured PR template with validation
  checklist — addresses `[MINOR] §10` (#141)

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

[Unreleased]: https://github.com/HomericIntelligence/ProjectArgus/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/HomericIntelligence/ProjectArgus/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/HomericIntelligence/ProjectArgus/releases/tag/v0.1.0
