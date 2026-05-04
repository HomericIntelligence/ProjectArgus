"""
Validate that all YAML config files parse correctly and have required top-level keys.
Uses only stdlib: yaml, pathlib, unittest.
"""
import unittest
import yaml
from pathlib import Path

REPO_ROOT = Path(__file__).parent.parent
CONFIGS_DIR = REPO_ROOT / "configs"


def load_yaml(path: Path) -> dict:
    with path.open() as f:
        return yaml.safe_load(f)


class TestPrometheusConfig(unittest.TestCase):
    def setUp(self):
        self.config = load_yaml(CONFIGS_DIR / "prometheus.yml")

    def test_parses_without_error(self):
        assert self.config is not None

    def test_has_global_section(self):
        assert "global" in self.config

    def test_global_has_scrape_interval(self):
        assert "scrape_interval" in self.config["global"]

    def test_global_has_evaluation_interval(self):
        assert "evaluation_interval" in self.config["global"]

    def test_has_scrape_configs(self):
        assert "scrape_configs" in self.config

    def test_scrape_configs_is_list(self):
        assert isinstance(self.config["scrape_configs"], list)

    def test_scrape_configs_not_empty(self):
        assert len(self.config["scrape_configs"]) > 0

    def test_each_scrape_config_has_job_name(self):
        for job in self.config["scrape_configs"]:
            assert "job_name" in job, f"Missing job_name in scrape config: {job}"

    def test_has_rule_files(self):
        assert "rule_files" in self.config


class TestLokiConfig(unittest.TestCase):
    def setUp(self):
        self.config = load_yaml(CONFIGS_DIR / "loki.yml")

    def test_parses_without_error(self):
        assert self.config is not None

    def test_has_server_section(self):
        assert "server" in self.config

    def test_server_has_http_listen_port(self):
        assert "http_listen_port" in self.config["server"]

    def test_has_schema_config(self):
        assert "schema_config" in self.config

    def test_schema_config_has_configs(self):
        assert "configs" in self.config["schema_config"]

    def test_has_limits_config(self):
        assert "limits_config" in self.config

    def test_limits_config_has_retention_period(self):
        assert "retention_period" in self.config["limits_config"]


class TestPromtailConfig(unittest.TestCase):
    def setUp(self):
        self.config = load_yaml(CONFIGS_DIR / "promtail.yml")

    def test_parses_without_error(self):
        assert self.config is not None

    def test_has_server_section(self):
        assert "server" in self.config

    def test_has_clients(self):
        assert "clients" in self.config

    def test_clients_is_list(self):
        assert isinstance(self.config["clients"], list)

    def test_clients_not_empty(self):
        assert len(self.config["clients"]) > 0

    def test_has_scrape_configs(self):
        assert "scrape_configs" in self.config

    def test_scrape_configs_is_list(self):
        assert isinstance(self.config["scrape_configs"], list)

    def test_syslog_job_host_label_uses_env_var(self):
        syslog_job = next(
            (j for j in self.config["scrape_configs"] if j.get("job_name") == "syslog"),
            None,
        )
        assert syslog_job is not None, "syslog scrape job not found"
        labels = syslog_job["static_configs"][0]["labels"]
        assert "host" in labels, "syslog job missing 'host' label"
        assert labels["host"].startswith("${"), (
            "host label must use env var substitution (${HOSTNAME:-...}), "
            f"got hardcoded value: {labels['host']!r}"
        )

    def test_syslog_job_host_label_has_fallback(self):
        syslog_job = next(
            (j for j in self.config["scrape_configs"] if j.get("job_name") == "syslog"),
            None,
        )
        assert syslog_job is not None
        host_val = syslog_job["static_configs"][0]["labels"]["host"]
        assert ":-" in host_val, (
            "host label env var should have a fallback default (e.g. ${HOSTNAME:-hermes}), "
            f"got: {host_val!r}"
        )


class TestGrafanaDatasourcesConfig(unittest.TestCase):
    def setUp(self):
        self.config = load_yaml(CONFIGS_DIR / "grafana" / "datasources.yml")

    def test_parses_without_error(self):
        assert self.config is not None

    def test_has_api_version(self):
        assert "apiVersion" in self.config

    def test_has_datasources(self):
        assert "datasources" in self.config

    def test_datasources_is_list(self):
        assert isinstance(self.config["datasources"], list)

    def test_datasources_not_empty(self):
        assert len(self.config["datasources"]) > 0

    def test_each_datasource_has_required_fields(self):
        required_fields = {"name", "type", "uid", "url"}
        for ds in self.config["datasources"]:
            for field in required_fields:
                assert field in ds, f"Datasource missing field '{field}': {ds}"


class TestGrafanaDashboardsConfig(unittest.TestCase):
    def setUp(self):
        self.config = load_yaml(CONFIGS_DIR / "grafana" / "dashboards.yml")

    def test_parses_without_error(self):
        assert self.config is not None

    def test_has_api_version(self):
        assert "apiVersion" in self.config

    def test_has_providers(self):
        assert "providers" in self.config

    def test_providers_is_list(self):
        assert isinstance(self.config["providers"], list)

    def test_providers_not_empty(self):
        assert len(self.config["providers"]) > 0

    def test_each_provider_has_required_fields(self):
        required_fields = {"name", "type", "options"}
        for provider in self.config["providers"]:
            for field in required_fields:
                assert field in provider, f"Provider missing field '{field}': {provider}"


if __name__ == "__main__":
    unittest.main()
