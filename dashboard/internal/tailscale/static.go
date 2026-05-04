package tailscale

import (
	"context"
	"os"
	"time"
)

// StaticSource is a Source that returns a fixed set of devices derived from
// environment variables. It is intended as a last-resort fallback when neither
// the CLI nor API sources are available.
//
// Environment variables:
//   - ATLAS_WORKER_HOST_IP  — IP for the "worker" device  (default: "127.0.0.1")
//   - ATLAS_CONTROL_HOST_IP — IP for the "control" device (default: "127.0.0.1")
type StaticSource struct{}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Devices always returns two devices with Online: true.
func (s StaticSource) Devices(_ context.Context) ([]Device, error) {
	workerIP := getEnvOrDefault("ATLAS_WORKER_HOST_IP", "127.0.0.1")
	controlIP := getEnvOrDefault("ATLAS_CONTROL_HOST_IP", "127.0.0.1")

	now := time.Now().UTC()
	return []Device{
		{
			Hostname:    "worker",
			TailscaleIP: workerIP,
			Online:      true,
			LastSeen:    now,
		},
		{
			Hostname:    "control",
			TailscaleIP: controlIP,
			Online:      true,
			LastSeen:    now,
		},
	}, nil
}
