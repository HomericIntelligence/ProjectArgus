"""Unit tests for exporter/exporter.py."""
from __future__ import annotations

import importlib
import io
import json
import sys
import types
import urllib.error
from io import BytesIO
from typing import Any
from unittest.mock import MagicMock, patch

import pytest

# ---------------------------------------------------------------------------
# Helpers to import the exporter module without it binding a port
# ---------------------------------------------------------------------------

def _import_exporter() -> types.ModuleType:
    """Import exporter.py cleanly, resetting the module each time."""
    if "exporter" in sys.modules:
        del sys.modules["exporter"]
    spec = importlib.util.spec_from_file_location(
        "exporter",
        "exporter/exporter.py",
    )
    assert spec and spec.loader
    mod = importlib.util.module_from_spec(spec)
    sys.modules["exporter"] = mod
    spec.loader.exec_module(mod)  # type: ignore[union-attr]
    return mod


@pytest.fixture()
def exporter():
    return _import_exporter()


# ---------------------------------------------------------------------------
# Fake HTTP response helper
# ---------------------------------------------------------------------------

def _fake_response(body: Any, status: int = 200) -> MagicMock:
    raw = json.dumps(body).encode() if not isinstance(body, bytes) else body
    mock = MagicMock()
    mock.status = status
    mock.read.return_value = raw
    mock.__enter__ = lambda s: s
    mock.__exit__ = MagicMock(return_value=False)
    return mock


# ---------------------------------------------------------------------------
# _fetch
# ---------------------------------------------------------------------------

class TestFetch:
    def test_success_returns_dict(self, exporter):
        payload = {"agents": []}
        with patch("urllib.request.urlopen", return_value=_fake_response(payload)):
            result = exporter._fetch("http://fake/v1/agents")
        assert result == payload

    def test_network_error_returns_none(self, exporter):
        with patch("urllib.request.urlopen", side_effect=OSError("timeout")):
            result = exporter._fetch("http://fake/v1/agents")
        assert result is None

    def test_json_decode_error_returns_none(self, exporter):
        mock = MagicMock()
        mock.read.return_value = b"not-json"
        mock.__enter__ = lambda s: s
        mock.__exit__ = MagicMock(return_value=False)
        with patch("urllib.request.urlopen", return_value=mock):
            result = exporter._fetch("http://fake/v1/agents")
        assert result is None


# ---------------------------------------------------------------------------
# _health_check
# ---------------------------------------------------------------------------

class TestHealthCheck:
    def test_http_200_returns_1(self, exporter):
        with patch("urllib.request.urlopen", return_value=_fake_response(b"ok", status=200)):
            assert exporter._health_check("http://fake/v1/health") == 1

    def test_http_500_returns_0(self, exporter):
        with patch("urllib.request.urlopen", return_value=_fake_response(b"err", status=500)):
            assert exporter._health_check("http://fake/v1/health") == 0

    def test_connection_refused_returns_0(self, exporter):
        with patch("urllib.request.urlopen", side_effect=OSError("refused")):
            assert exporter._health_check("http://fake/v1/health") == 0


# ---------------------------------------------------------------------------
# collect()
# ---------------------------------------------------------------------------

_AGENTS_RESPONSE = {
    "agents": [
        {"name": "alpha", "host": "host1", "program": "nestor", "status": "online"},
        {"name": "beta",  "host": "host2", "program": "hermes", "status": "offline"},
    ]
}

_TASKS_RESPONSE = {
    "tasks": [
        {"status": "completed"},
        {"status": "completed"},
        {"status": "failed"},
    ]
}

_NESTOR_STATS_RESPONSE = {
    "active": 3,
    "completed": 10,
    "pending": 1,
}

_VARZ_RESPONSE = {
    "connections": 5,
    "in_msgs": 1000,
    "out_msgs": 900,
    "in_bytes": 50000,
    "out_bytes": 45000,
    "slow_consumers": 0,
}

_JSZ_RESPONSE = {
    "streams": 2,
    "consumers": 4,
    "messages": 5000,
    "bytes": 200000,
}


def _url_dispatch(url: str, **kwargs: object) -> MagicMock:
    """Return a fake HTTP response based on URL path."""
    if "/v1/health" in url:
        return _fake_response(b"ok", status=200)
    if "/v1/agents" in url:
        return _fake_response(_AGENTS_RESPONSE)
    if "/v1/tasks" in url:
        return _fake_response(_TASKS_RESPONSE)
    if "/v1/research/stats" in url:
        return _fake_response(_NESTOR_STATS_RESPONSE)
    if "/varz" in url:
        return _fake_response(_VARZ_RESPONSE)
    if "/jsz" in url:
        return _fake_response(_JSZ_RESPONSE)
    raise ValueError(f"unexpected URL: {url}")


@pytest.fixture()
def metrics_output(exporter) -> str:
    with patch("urllib.request.urlopen", side_effect=_url_dispatch):
        return exporter.collect()


class TestCollect:
    def test_hi_agents_total(self, metrics_output):
        assert "hi_agents_total 2" in metrics_output

    def test_hi_agents_online(self, metrics_output):
        assert "hi_agents_online 1" in metrics_output

    def test_hi_agents_offline(self, metrics_output):
        assert "hi_agents_offline 1" in metrics_output

    def test_hi_agamemnon_health(self, metrics_output):
        assert "hi_agamemnon_health 1" in metrics_output

    def test_hi_nestor_health(self, metrics_output):
        assert "hi_nestor_health 1" in metrics_output

    def test_hi_tasks_total(self, metrics_output):
        assert "hi_tasks_total 3" in metrics_output

    def test_hi_tasks_by_status_completed(self, metrics_output):
        assert 'hi_tasks_by_status{status="completed"} 2' in metrics_output

    def test_hi_tasks_by_status_failed(self, metrics_output):
        assert 'hi_tasks_by_status{status="failed"} 1' in metrics_output

    def test_nats_connections(self, metrics_output):
        assert "nats_connections 5" in metrics_output

    def test_nats_in_msgs_total(self, metrics_output):
        assert "nats_in_msgs_total 1000" in metrics_output

    def test_nats_jetstream_bytes(self, metrics_output):
        assert "nats_jetstream_bytes 200000" in metrics_output

    def test_scrape_timestamp_present(self, metrics_output):
        assert "homeric_exporter_scrape_timestamp" in metrics_output

    def test_no_duplicate_type_lines(self, metrics_output):
        """Each metric name must appear in a # TYPE line exactly once."""
        type_counts: dict[str, int] = {}
        for line in metrics_output.splitlines():
            if line.startswith("# TYPE "):
                name = line.split()[2]
                type_counts[name] = type_counts.get(name, 0) + 1
        duplicates = {k: v for k, v in type_counts.items() if v > 1}
        assert duplicates == {}, f"Duplicate # TYPE declarations: {duplicates}"

    def test_help_lines_present(self, metrics_output):
        assert "# HELP hi_agents_total" in metrics_output
        assert "# HELP nats_connections" in metrics_output
        assert "# HELP homeric_exporter_scrape_timestamp" in metrics_output

    def test_per_agent_label(self, metrics_output):
        assert 'hi_agent_online{name="alpha"' in metrics_output
        assert 'hi_agent_online{name="beta"' in metrics_output

    def test_nestor_research_stats(self, metrics_output):
        assert "hi_nestor_research_active 3" in metrics_output
        assert "hi_nestor_research_completed 10" in metrics_output
        assert "hi_nestor_research_pending 1" in metrics_output

    def test_collect_with_all_endpoints_down(self, exporter):
        """collect() must not raise when all upstream services are unreachable."""
        with patch("urllib.request.urlopen", side_effect=OSError("unreachable")):
            output = exporter.collect()
        assert "hi_agamemnon_health 0" in output
        assert "hi_nestor_health 0" in output
        assert "homeric_exporter_scrape_timestamp" in output


# ---------------------------------------------------------------------------
# Handler HTTP responses
# ---------------------------------------------------------------------------

class _FakeStream(io.BytesIO):
    """BytesIO that captures HTTP response bytes."""
    pass


def _make_handler(exporter_mod, path: str) -> tuple[Any, _FakeStream]:
    """Instantiate Handler for a GET <path> request, return (handler, wfile).

    do_GET only uses self.path and self.wfile so we skip the full HTTP parse.
    """
    output = _FakeStream()

    handler = exporter_mod.Handler.__new__(exporter_mod.Handler)
    handler.client_address = ("127.0.0.1", 12345)
    handler.server = MagicMock()
    handler.request = MagicMock()
    handler.rfile = io.BytesIO(b"")
    handler.wfile = output
    handler.requestline = f"GET {path} HTTP/1.1"
    handler.command = "GET"
    handler.path = path
    handler.request_version = "HTTP/1.1"
    handler.headers = MagicMock()
    handler.headers.get = MagicMock(return_value=None)
    return handler, output


class TestHandler:
    def test_metrics_returns_200(self, exporter):
        handler, output = _make_handler(exporter, "/metrics")
        with patch("urllib.request.urlopen", side_effect=_url_dispatch):
            handler.do_GET()
        response = output.getvalue().decode()
        assert "200 OK" in response
        assert "hi_agents_total" in response

    def test_health_returns_ok(self, exporter):
        handler, output = _make_handler(exporter, "/health")
        handler.do_GET()
        response = output.getvalue().decode()
        assert "200 OK" in response
        assert "ok" in response

    def test_unknown_path_returns_404(self, exporter):
        handler, output = _make_handler(exporter, "/unknown")
        handler.do_GET()
        response = output.getvalue().decode()
        assert "404" in response
