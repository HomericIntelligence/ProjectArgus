#!/usr/bin/env python3
"""
homeric-exporter — Converts Agamemnon, Nestor, and NATS JSON APIs to Prometheus metrics.
Runs as a sidecar in the argus stack, exposes /metrics on port 9100.
"""
from __future__ import annotations

import json
import logging
import os
import ssl
import time
import urllib.request
from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Optional

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger("homeric-exporter")

AGAMEMNON_URL     = os.environ.get("AGAMEMNON_URL",     "http://172.20.0.1:8080")
NESTOR_URL        = os.environ.get("NESTOR_URL",        "http://172.20.0.1:8081")
NATS_URL          = os.environ.get("NATS_URL",          "http://172.24.0.1:8222")
PORT              = int(os.environ.get("EXPORTER_PORT", "9100"))

# Optional CA bundle paths for TLS verification on each upstream.
# Set to the path of a CA certificate file (PEM) to enable custom trust.
# Leave unset to use the system trust store (appropriate when the upstream
# uses a publicly-trusted cert or when Tailscale handles transport encryption).
AGAMEMNON_TLS_CA  = os.environ.get("AGAMEMNON_TLS_CA")
NESTOR_TLS_CA     = os.environ.get("NESTOR_TLS_CA")
NATS_TLS_CA       = os.environ.get("NATS_TLS_CA")

# Set TLS_VERIFY=false to disable certificate verification entirely.
# Only for development — never disable in production.
_TLS_VERIFY       = os.environ.get("TLS_VERIFY", "true").lower() != "false"


def _build_ssl_context(ca_file: Optional[str] = None) -> Optional[ssl.SSLContext]:
    """Return an SSLContext for HTTPS requests, or None for plain HTTP."""
    if not _TLS_VERIFY:
        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE
        return ctx
    if ca_file:
        ctx = ssl.create_default_context(cafile=ca_file)
        return ctx
    # No custom CA specified; use the system trust store (default urllib behaviour).
    return None


def _fetch(url: str, ca_file: Optional[str] = None) -> dict | None:
    try:
        ctx = _build_ssl_context(ca_file)
        r = urllib.request.urlopen(url, timeout=5, context=ctx)
        return json.loads(r.read())
    except Exception as e:
        log.warning("fetch %s failed: %s", url, e)
        return None


def _health_check(url: str, ca_file: Optional[str] = None) -> int:
    """Return 1 if the URL returns HTTP 200, 0 otherwise."""
    try:
        ctx = _build_ssl_context(ca_file)
        r = urllib.request.urlopen(url, timeout=5, context=ctx)
        return 1 if r.status == 200 else 0
    except Exception:
        return 0


def collect() -> str:
    lines: list[str] = []

    def gauge(name: str, value: float | int, labels: dict = {}) -> None:
        lstr = ",".join(f'{k}="{v}"' for k, v in labels.items())
        lines.append(f"# TYPE {name} gauge")
        lines.append(f"{name}{{{lstr}}} {value}")

    # ── Agamemnon health ───────────────────────────────────────────────────
    gauge("hi_agamemnon_health", _health_check(f"{AGAMEMNON_URL}/v1/health", AGAMEMNON_TLS_CA))

    # ── Agamemnon agents ───────────────────────────────────────────────────
    d = _fetch(f"{AGAMEMNON_URL}/v1/agents", AGAMEMNON_TLS_CA)
    if d:
        agents = d.get("agents", [])
        total   = len(agents)
        online  = sum(1 for a in agents if a.get("status") == "online")
        offline = total - online
        gauge("hi_agents_total",   total)
        gauge("hi_agents_online",  online)
        gauge("hi_agents_offline", offline)
        for ag in agents:
            gauge("hi_agent_online",
                  1 if ag.get("status") == "online" else 0,
                  {"name":    ag.get("name", "unknown"),
                   "host":    ag.get("host", "unknown"),
                   "program": ag.get("program", "unknown")})

    # ── Agamemnon tasks ────────────────────────────────────────────────────
    d = _fetch(f"{AGAMEMNON_URL}/v1/tasks", AGAMEMNON_TLS_CA)
    if d:
        tasks = d.get("tasks", [])
        gauge("hi_tasks_total", len(tasks))
        status_counts: dict[str, int] = {}
        for task in tasks:
            s = task.get("status", "unknown")
            status_counts[s] = status_counts.get(s, 0) + 1
        for status, count in status_counts.items():
            gauge("hi_tasks_by_status", count, {"status": status})

    # ── Nestor health + research stats ────────────────────────────────────
    gauge("hi_nestor_health", _health_check(f"{NESTOR_URL}/v1/health", NESTOR_TLS_CA))

    d = _fetch(f"{NESTOR_URL}/v1/research/stats", NESTOR_TLS_CA)
    if d:
        gauge("hi_nestor_research_active",    d.get("active", 0))
        gauge("hi_nestor_research_completed", d.get("completed", 0))
        gauge("hi_nestor_research_pending",   d.get("pending", 0))

    # ── NATS ───────────────────────────────────────────────────────────────
    d = _fetch(f"{NATS_URL}/varz", NATS_TLS_CA)
    if d:
        gauge("nats_connections",    d.get("connections", 0))
        gauge("nats_in_msgs_total",  d.get("in_msgs", 0))
        gauge("nats_out_msgs_total", d.get("out_msgs", 0))
        gauge("nats_in_bytes_total", d.get("in_bytes", 0))
        gauge("nats_out_bytes_total",d.get("out_bytes", 0))
        gauge("nats_slow_consumers", d.get("slow_consumers", 0))

    d = _fetch(f"{NATS_URL}/jsz", NATS_TLS_CA)
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

    def log_message(self, fmt, *args):
        pass


if __name__ == "__main__":
    log.info("homeric-exporter starting on port %d", PORT)
    log.info("Scraping Agamemnon at %s (CA: %s)", AGAMEMNON_URL, AGAMEMNON_TLS_CA or "system trust store")
    log.info("Scraping Nestor at %s (CA: %s)", NESTOR_URL, NESTOR_TLS_CA or "system trust store")
    log.info("Scraping NATS at %s (CA: %s)", NATS_URL, NATS_TLS_CA or "system trust store")
    if not _TLS_VERIFY:
        log.warning("TLS certificate verification is DISABLED (TLS_VERIFY=false)")
    HTTPServer(("0.0.0.0", PORT), Handler).serve_forever()
