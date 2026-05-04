"""
Unit tests for exporter/exporter.py.

All network calls are mocked via unittest.mock.patch so no real HTTP
connections are made during the test suite.
"""
from __future__ import annotations

import io
import json
import sys
import unittest
import urllib.error
from pathlib import Path
from unittest.mock import MagicMock, patch

# Make the exporter importable without running __main__ logic
REPO_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(REPO_ROOT))
import exporter.exporter as exporter_mod  # noqa: E402


def _make_response(data: dict | None = None, status: int = 200) -> MagicMock:
    """Return a mock object that behaves like urllib.request.urlopen's return value."""
    mock = MagicMock()
    mock.status = status
    if data is not None:
        mock.read.return_value = json.dumps(data).encode()
    else:
        mock.read.return_value = b"{}"
    mock.__enter__ = lambda s: s
    mock.__exit__ = MagicMock(return_value=False)
    return mock


def _urlopen_raises(*args, **kwargs):
    raise OSError("connection refused")


# ---------------------------------------------------------------------------
# Test _health_check
# ---------------------------------------------------------------------------

class TestHealthCheck(unittest.TestCase):
    def test_returns_1_for_http_200(self):
        mock_resp = _make_response(status=200)
        with patch("urllib.request.urlopen", return_value=mock_resp):
            result = exporter_mod._health_check("http://fake/health")
        self.assertEqual(result, 1)

    def test_returns_0_for_non_200(self):
        mock_resp = _make_response(status=503)
        with patch("urllib.request.urlopen", return_value=mock_resp):
            result = exporter_mod._health_check("http://fake/health")
        self.assertEqual(result, 0)

    def test_returns_0_on_exception(self):
        with patch("urllib.request.urlopen", side_effect=_urlopen_raises):
            result = exporter_mod._health_check("http://fake/health")
        self.assertEqual(result, 0)


# ---------------------------------------------------------------------------
# Test _fetch
# ---------------------------------------------------------------------------

class TestFetch(unittest.TestCase):
    def test_returns_dict_on_success(self):
        mock_resp = _make_response({"key": "value"})
        with patch("urllib.request.urlopen", return_value=mock_resp):
            result = exporter_mod._fetch("http://fake/data")
        self.assertIsInstance(result, dict)
        self.assertEqual(result["key"], "value")

    def test_returns_none_on_oserror(self):
        with patch("urllib.request.urlopen", side_effect=OSError("connection refused")):
            result = exporter_mod._fetch("http://fake/data")
        self.assertIsNone(result)

    def test_returns_none_on_urlerror(self):
        with patch("urllib.request.urlopen",
                   side_effect=urllib.error.URLError("name or service not known")):
            result = exporter_mod._fetch("http://fake/data")
        self.assertIsNone(result)

    def test_returns_none_on_json_decode_error(self):
        mock_resp = MagicMock()
        mock_resp.read.return_value = b"not-json"
        mock_resp.__enter__ = lambda s: s
        mock_resp.__exit__ = MagicMock(return_value=False)
        with patch("urllib.request.urlopen", return_value=mock_resp):
            result = exporter_mod._fetch("http://fake/data")
        self.assertIsNone(result)

    def test_propagates_unexpected_exception(self):
        """Exceptions outside the specific tuple must not be swallowed."""
        with patch("urllib.request.urlopen", side_effect=MemoryError("oom")):
            with self.assertRaises(MemoryError):
                exporter_mod._fetch("http://fake/data")

    def test_returns_none_on_exception(self):
        with patch("urllib.request.urlopen", side_effect=_urlopen_raises):
            result = exporter_mod._fetch("http://fake/data")
        self.assertIsNone(result)


# ---------------------------------------------------------------------------
# Helper: patch all seven upstream calls in collect()
# ---------------------------------------------------------------------------

def _patch_collect(
    agamemnon_health: int = 1,
    agents_data: dict | None = None,
    tasks_data: dict | None = None,
    nestor_health: int = 1,
    nestor_stats: dict | None = None,
    nats_varz: dict | None = None,
    nats_jsz: dict | None = None,
):
    """Context-manager factory that patches _health_check and _fetch inside collect()."""
    agents_data = agents_data or {}
    tasks_data = tasks_data or {}

    def _fake_health_check(url: str) -> int:
        if "agamemnon" in url or "8080" in url:
            return agamemnon_health
        return nestor_health

    def _fake_fetch(url: str) -> dict | None:
        if "/v1/agents" in url:
            return agents_data
        if "/v1/tasks" in url:
            return tasks_data
        if "/research/stats" in url:
            return nestor_stats
        if "/varz" in url:
            return nats_varz
        if "/jsz" in url:
            return nats_jsz
        return None

    return (
        patch.object(exporter_mod, "_health_check", side_effect=_fake_health_check),
        patch.object(exporter_mod, "_fetch", side_effect=_fake_fetch),
    )


# ---------------------------------------------------------------------------
# Test collect() — output format
# ---------------------------------------------------------------------------

class TestCollectFormat(unittest.TestCase):
    def _run_collect(self, **kwargs):
        hc_patch, fetch_patch = _patch_collect(**kwargs)
        with hc_patch, fetch_patch:
            return exporter_mod.collect()

    def test_returns_string(self):
        output = self._run_collect()
        self.assertIsInstance(output, str)

    def test_ends_with_newline(self):
        output = self._run_collect()
        self.assertTrue(output.endswith("\n"), "collect() output must end with newline")

    def test_contains_type_declarations(self):
        output = self._run_collect()
        self.assertIn("# TYPE", output, "output must contain at least one # TYPE declaration")

    def test_no_exception_when_all_upstreams_down(self):
        """collect() must not raise even if every upstream returns None."""
        hc_patch, fetch_patch = _patch_collect(
            agamemnon_health=0,
            agents_data=None,
            tasks_data=None,
            nestor_health=0,
            nestor_stats=None,
            nats_varz=None,
            nats_jsz=None,
        )
        try:
            with hc_patch, fetch_patch:
                output = exporter_mod.collect()
        except Exception as exc:
            self.fail(f"collect() raised an exception when all upstreams are down: {exc}")
        self.assertIsInstance(output, str)

    def test_type_emitted_once_per_metric(self):
        """Each metric name must have exactly one # TYPE line (no duplicates)."""
        nats_varz = {
            "connections": 3, "in_msgs": 100, "out_msgs": 90,
            "in_bytes": 1024, "out_bytes": 512, "slow_consumers": 0,
        }
        output = self._run_collect(nats_varz=nats_varz)
        type_lines = [line for line in output.splitlines() if line.startswith("# TYPE")]
        names = [line.split()[2] for line in type_lines]
        self.assertEqual(len(names), len(set(names)),
                         "Duplicate # TYPE declarations found in collect() output")


# ---------------------------------------------------------------------------
# Test collect() — metric names and values
# ---------------------------------------------------------------------------

class TestCollectMetricNames(unittest.TestCase):
    def setUp(self):
        self.agents_data = {
            "agents": [
                {"name": "alpha", "host": "h1", "program": "prog", "status": "online"},
                {"name": "beta",  "host": "h2", "program": "prog", "status": "offline"},
            ]
        }
        self.tasks_data = {
            "tasks": [
                {"status": "completed"},
                {"status": "completed"},
                {"status": "failed"},
            ]
        }
        self.nats_varz = {
            "connections": 5, "in_msgs": 200, "out_msgs": 180,
            "in_bytes": 2048, "out_bytes": 1024, "slow_consumers": 1,
        }
        self.nestor_stats = {"active": 2, "completed": 10, "pending": 1}
        hc_patch, fetch_patch = _patch_collect(
            agamemnon_health=1,
            agents_data=self.agents_data,
            tasks_data=self.tasks_data,
            nestor_health=1,
            nestor_stats=self.nestor_stats,
            nats_varz=self.nats_varz,
        )
        with hc_patch, fetch_patch:
            self.output = exporter_mod.collect()

    def test_contains_agamemnon_health(self):
        self.assertIn("hi_agamemnon_health", self.output)

    def test_contains_nestor_health(self):
        self.assertIn("hi_nestor_health", self.output)

    def test_contains_nats_connections(self):
        self.assertIn("nats_connections", self.output)

    def test_agent_totals_correct(self):
        """hi_agents_total, hi_agents_online, hi_agents_offline values."""
        lines = {ln.split()[0]: ln.split()[1]
                 for ln in self.output.splitlines()
                 if not ln.startswith("#") and ln.strip()}
        self.assertEqual(lines.get("hi_agents_total{}"), "2")
        self.assertEqual(lines.get("hi_agents_online{}"), "1")
        self.assertEqual(lines.get("hi_agents_offline{}"), "1")

    def test_task_total_correct(self):
        lines = {ln.split()[0]: ln.split()[1]
                 for ln in self.output.splitlines()
                 if not ln.startswith("#") and ln.strip()}
        self.assertEqual(lines.get("hi_tasks_total{}"), "3")

    def test_exporter_self_metrics_present(self):
        self.assertIn("homeric_exporter_scrape_duration_seconds", self.output)
        self.assertIn("homeric_exporter_scrape_timestamp", self.output)
        self.assertIn("homeric_exporter_fetch_errors_total", self.output)


# ---------------------------------------------------------------------------
# Test Handler (HTTP server)
# ---------------------------------------------------------------------------

class _FakeSocket:
    """Minimal socket-like object for BaseHTTPRequestHandler tests."""
    def __init__(self, request_bytes: bytes):
        self._data = request_bytes
        self._wfile = io.BytesIO()

    def makefile(self, mode, **kwargs):
        if "r" in mode:
            return io.BufferedReader(io.BytesIO(self._data))
        return self._wfile

    def sendall(self, data: bytes):
        self._wfile.write(data)


def _make_handler(path: str) -> tuple[exporter_mod.Handler, io.BytesIO]:
    """Instantiate Handler for a fake GET request to *path*, return (handler, wfile)."""
    request_bytes = f"GET {path} HTTP/1.1\r\nHost: localhost\r\n\r\n".encode()
    fake_socket = _FakeSocket(request_bytes)
    server = MagicMock()
    # BaseHTTPRequestHandler.__init__ calls setup() then handle() then finish()
    # We call the constructor which invokes handle() -> do_GET()
    # We need to suppress that here; instead we'll call do_GET manually after setup.
    handler = exporter_mod.Handler.__new__(exporter_mod.Handler)
    handler.rfile = io.BufferedReader(io.BytesIO(request_bytes))
    handler.wfile = fake_socket._wfile
    handler.client_address = ("127.0.0.1", 9999)
    handler.server = server
    handler.path = path
    handler.request_version = "HTTP/1.1"
    handler.command = "GET"
    handler.requestline = f"GET {path} HTTP/1.1"
    handler.headers = {}
    handler.connection = MagicMock()
    return handler, fake_socket._wfile


class TestHandler(unittest.TestCase):
    def _get_response(self, path: str, mock_collect_output: str = "# TYPE x gauge\nx{} 1\n"):
        handler, wfile = _make_handler(path)
        with patch.object(exporter_mod, "collect", return_value=mock_collect_output):
            handler.do_GET()
        wfile.seek(0)
        return wfile.read().decode(errors="replace")

    def test_health_returns_200(self):
        response = self._get_response("/health")
        self.assertIn("200", response)

    def test_health_body_is_ok(self):
        response = self._get_response("/health")
        self.assertIn("ok", response)

    def test_metrics_returns_200(self):
        response = self._get_response("/metrics")
        self.assertIn("200", response)

    def test_metrics_content_type(self):
        response = self._get_response("/metrics")
        self.assertIn("text/plain", response)
        self.assertIn("version=0.0.4", response)

    def test_unknown_path_returns_404(self):
        response = self._get_response("/notfound")
        self.assertIn("404", response)

    def test_metrics_body_contains_collect_output(self):
        collect_output = "# TYPE hi_agents_total gauge\nhi_agents_total{} 42\n"
        response = self._get_response("/metrics", mock_collect_output=collect_output)
        self.assertIn("hi_agents_total", response)


if __name__ == "__main__":
    unittest.main()
