package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr          string
	LogLevel            slog.Level
	NATSURL             string
	NATSMonURL          string
	AgamemnonURL        string
	NestorURL           string
	HermesURL           string
	PrometheusURL       string
	GrafanaURL          string
	LokiURL             string
	ExporterURL         string
	MnemosyneSkillsDir  string
	TailscaleSource     string
	TailscaleAPIKey     string
	TailnetName         string
	TailscaleSocket     string
	AuthMode            string
	AuthUser            string
	AuthPass            string
	AuthBearerToken     string
	PollAgamemnonMs     time.Duration
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() *Config {
	logLevelStr := getenv("ATLAS_LOG_LEVEL", "info")
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(logLevelStr)); err != nil {
		logLevel = slog.LevelInfo
	}

	pollMs, _ := strconv.Atoi(getenv("ATLAS_POLL_AGAMEMNON_MS", "5000"))

	return &Config{
		ListenAddr:         getenv("ATLAS_LISTEN_ADDR", ":3002"),
		LogLevel:           logLevel,
		NATSURL:            getenv("ATLAS_NATS_URL", "nats://nats:4222"),
		NATSMonURL:         getenv("ATLAS_NATS_MON_URL", "http://nats:8222"),
		AgamemnonURL:       getenv("ATLAS_AGAMEMNON_URL", "http://agamemnon:8080"),
		NestorURL:          getenv("ATLAS_NESTOR_URL", "http://nestor:8081"),
		HermesURL:          getenv("ATLAS_HERMES_URL", "http://hermes:8085"),
		PrometheusURL:      getenv("ATLAS_PROMETHEUS_URL", "http://prometheus:9090"),
		GrafanaURL:         getenv("ATLAS_GRAFANA_URL", "http://grafana:3000"),
		LokiURL:            getenv("ATLAS_LOKI_URL", "http://loki:3100"),
		ExporterURL:        getenv("ATLAS_EXPORTER_URL", "http://argus-exporter:9100"),
		MnemosyneSkillsDir: getenv("ATLAS_MNEMOSYNE_SKILLS_DIR", "/mnt/mnemosyne/skills"),
		TailscaleSource:    getenv("ATLAS_TAILSCALE_SOURCE", "static"),
		TailscaleAPIKey:    getenv("ATLAS_TAILSCALE_API_KEY", ""),
		TailnetName:        getenv("ATLAS_TAILNET_NAME", ""),
		TailscaleSocket:    getenv("ATLAS_TAILSCALE_SOCKET", "/var/run/tailscale/tailscaled.sock"),
		AuthMode:           getenv("ATLAS_AUTH_MODE", "none"),
		AuthUser:           getenv("ATLAS_AUTH_USER", ""),
		AuthPass:           getenv("ATLAS_AUTH_PASS", ""),
		AuthBearerToken:    getenv("ATLAS_AUTH_BEARER_TOKEN", ""),
		PollAgamemnonMs:    time.Duration(pollMs) * time.Millisecond,
	}
}
