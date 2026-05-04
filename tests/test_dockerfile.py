"""
Validate that exporter/Dockerfile meets production hardening requirements:
non-root user, embedded HEALTHCHECK, and WORKDIR setup.
"""
from __future__ import annotations

import re
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).parent.parent
DOCKERFILE = REPO_ROOT / "exporter" / "Dockerfile"


def _lines() -> list[str]:
    return DOCKERFILE.read_text().splitlines()


class TestDockerfileExists(unittest.TestCase):
    def test_dockerfile_present(self) -> None:
        self.assertTrue(DOCKERFILE.exists(), "exporter/Dockerfile not found")


class TestDockerfileNonRootUser(unittest.TestCase):
    def test_has_user_directive(self) -> None:
        lines = _lines()
        user_lines = [ln for ln in lines if re.match(r"^USER\s+", ln)]
        self.assertTrue(user_lines, "Dockerfile must have a USER directive")

    def test_user_is_not_root(self) -> None:
        lines = _lines()
        for ln in lines:
            if re.match(r"^USER\s+", ln):
                value = ln.split(None, 1)[1].strip()
                self.assertNotIn(value, {"0", "root"}, "USER must not be root (0)")

    def test_creates_non_root_group_and_user(self) -> None:
        content = DOCKERFILE.read_text()
        self.assertIn("groupadd", content, "Dockerfile must create a system group")
        self.assertIn("useradd", content, "Dockerfile must create a system user")

    def test_user_directive_after_copy(self) -> None:
        lines = _lines()
        copy_idx = next((i for i, ln in enumerate(lines) if ln.startswith("COPY")), None)
        user_idx = next((i for i, ln in enumerate(lines) if re.match(r"^USER\s+", ln)), None)
        self.assertIsNotNone(copy_idx, "Dockerfile must have a COPY directive")
        self.assertIsNotNone(user_idx, "Dockerfile must have a USER directive")
        self.assertGreater(user_idx, copy_idx, "USER directive must come after COPY")


class TestDockerfileHealthcheck(unittest.TestCase):
    def test_has_healthcheck(self) -> None:
        content = DOCKERFILE.read_text()
        self.assertIn("HEALTHCHECK", content, "Dockerfile must embed a HEALTHCHECK")

    def test_healthcheck_not_none(self) -> None:
        content = DOCKERFILE.read_text()
        self.assertNotIn("HEALTHCHECK NONE", content, "HEALTHCHECK must not be disabled")

    def test_healthcheck_uses_python(self) -> None:
        content = DOCKERFILE.read_text()
        self.assertIn("python", content.lower(),
                      "HEALTHCHECK probe must use python (no wget/curl in slim image)")

    def test_healthcheck_hits_health_endpoint(self) -> None:
        content = DOCKERFILE.read_text()
        self.assertIn("/health", content, "HEALTHCHECK must probe the /health endpoint")

    def test_healthcheck_has_interval(self) -> None:
        content = DOCKERFILE.read_text()
        self.assertIn("--interval=", content, "HEALTHCHECK must specify --interval")

    def test_healthcheck_has_timeout(self) -> None:
        content = DOCKERFILE.read_text()
        self.assertIn("--timeout=", content, "HEALTHCHECK must specify --timeout")


class TestDockerfileWorkdir(unittest.TestCase):
    def test_has_workdir(self) -> None:
        lines = _lines()
        workdir_lines = [ln for ln in lines if ln.startswith("WORKDIR")]
        self.assertTrue(workdir_lines, "Dockerfile must have a WORKDIR directive")

    def test_cmd_references_workdir_path(self) -> None:
        lines = _lines()
        workdir_val = next(
            (ln.split(None, 1)[1].strip() for ln in lines if ln.startswith("WORKDIR")),
            None,
        )
        cmd_lines = [ln for ln in lines if ln.startswith("CMD")]
        self.assertTrue(cmd_lines, "Dockerfile must have a CMD directive")
        if workdir_val:
            # CMD should reference the script under the WORKDIR path
            self.assertTrue(
                any(workdir_val in ln for ln in cmd_lines),
                f"CMD must reference the script path under WORKDIR ({workdir_val})",
            )


class TestDockerfileChown(unittest.TestCase):
    def test_copy_uses_chown(self) -> None:
        content = DOCKERFILE.read_text()
        self.assertIn("--chown=", content, "COPY must use --chown to set file ownership")


if __name__ == "__main__":
    unittest.main()
