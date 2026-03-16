#!/usr/bin/env python3
"""
homeric-exporter — Converts ai-maestro and NATS JSON APIs to Prometheus metrics.
Runs as a sidecar in the argus stack, exposes /metrics on port 9100.
"""
from __future__ import annotations

import os
import time
import urllib.request
import json
import logging
from http.server import BaseHTTPRequestHandler, HTTPServer

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger("homeric-exporter")

MAESTRO_URL = os.environ.get("MAESTRO_URL", "http://172.20.0.1:23000")
NATS_URL    = os.environ.get("NATS_URL",    "http://172.24.0.1:8222")
PORT        = int(os.environ.get("EXPORTER_PORT", "9100"))


def _fetch(url: str) -> dict | None:
    try:
        r = urllib.request.urlopen(url, timeout=5)
        return json.loads(r.read())
    except Exception as e:
        log.warning("fetch %s failed: %s", url, e)
        return None


def collect() -> str:
    lines: list[str] = []

    def gauge(name: str, value: float | int, labels: dict = {}) -> None:
        lstr = ",".join(f'{k}="{v}"' for k, v in labels.items())
        lines.append(f"# TYPE {name} gauge")
        lines.append(f"{name}{{{lstr}}} {value}")

    # ── ai-maestro agents ──────────────────────────────────────────────────
    d = _fetch(f"{MAESTRO_URL}/api/agents/unified")
    if d:
        stats = d.get("stats", {})
        gauge("maestro_agents_total",      stats.get("total", 0))
        gauge("maestro_agents_online",     stats.get("online", 0))
        gauge("maestro_agents_offline",    stats.get("offline", 0))
        gauge("maestro_agents_orphans",    stats.get("orphans", 0))
        gauge("maestro_hosts_total",       d.get("totalHosts", 0))
        gauge("maestro_hosts_successful",  d.get("successfulHosts", 0))
        for entry in d.get("agents", []):
            ag = entry["agent"]
            online = 1 if ag.get("session", {}).get("status") == "online" else 0
            gauge("maestro_agent_online", online, {
                "name": ag["name"],
                "host": ag.get("hostId", "unknown"),
                "program": ag.get("program", "unknown"),
            })

    # ── ai-maestro diagnostics ─────────────────────────────────────────────
    d = _fetch(f"{MAESTRO_URL}/api/diagnostics")
    if d:
        summary = d.get("summary", {})
        gauge("maestro_diagnostics_total",   summary.get("total", 0))
        gauge("maestro_diagnostics_passed",  summary.get("passed", 0))
        gauge("maestro_diagnostics_failed",  summary.get("failed", 0))
        gauge("maestro_diagnostics_ok",      1 if summary.get("status") == "pass" else 0)
        for check in d.get("checks", []):
            gauge("maestro_check_ok", 1 if check["status"] == "pass" else 0,
                  {"check": check["name"]})

    # ── ai-maestro teams + tasks ───────────────────────────────────────────
    d = _fetch(f"{MAESTRO_URL}/api/teams")
    if d:
        teams = d.get("teams", [])
        gauge("maestro_teams_total", len(teams))
        total_tasks = 0
        status_counts: dict[str, int] = {}
        for team in teams:
            td = _fetch(f"{MAESTRO_URL}/api/teams/{team['id']}/tasks")
            if td:
                for task in td.get("tasks", []):
                    total_tasks += 1
                    s = task.get("status", "unknown")
                    status_counts[s] = status_counts.get(s, 0) + 1
        gauge("maestro_tasks_total", total_tasks)
        for status, count in status_counts.items():
            gauge("maestro_tasks_by_status", count, {"status": status})

    # ── NATS ───────────────────────────────────────────────────────────────
    d = _fetch(f"{NATS_URL}/varz")
    if d:
        gauge("nats_connections",     d.get("connections", 0))
        gauge("nats_in_msgs_total",   d.get("in_msgs", 0))
        gauge("nats_out_msgs_total",  d.get("out_msgs", 0))
        gauge("nats_in_bytes_total",  d.get("in_bytes", 0))
        gauge("nats_out_bytes_total", d.get("out_bytes", 0))
        gauge("nats_slow_consumers",  d.get("slow_consumers", 0))

    d = _fetch(f"{NATS_URL}/jsz")
    if d:
        gauge("nats_jetstream_streams",   d.get("streams", 0))
        gauge("nats_jetstream_consumers", d.get("consumers", 0))
        gauge("nats_jetstream_messages",  d.get("messages", 0))
        gauge("nats_jetstream_bytes",     d.get("bytes", 0))

    # ── exporter self ──────────────────────────────────────────────────────
    gauge("homeric_exporter_scrape_timestamp", time.time())

    return "\n".join(lines) + "\n"


class Handler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        if self.path == "/metrics":
            body = collect().encode()
            self.send_response(200)
            self.send_header("Content-Type", "text/plain; version=0.0.4")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
        elif self.path == "/health":
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b"ok")
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, fmt, *args):  # suppress per-request access log noise
        pass


if __name__ == "__main__":
    log.info("homeric-exporter starting on port %d", PORT)
    log.info("Scraping ai-maestro at %s", MAESTRO_URL)
    log.info("Scraping NATS at %s", NATS_URL)
    HTTPServer(("0.0.0.0", PORT), Handler).serve_forever()
