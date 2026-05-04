package tailscale

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// cliStatus mirrors the JSON output of `tailscale status --json`.
type cliStatus struct {
	Self  cliPeer            `json:"Self"`
	Peers map[string]cliPeer `json:"Peer"`
}

type cliPeer struct {
	HostName     string   `json:"HostName"`
	TailscaleIPs []string `json:"TailscaleIPs"`
	Online       bool     `json:"Online"`
	LastSeen     string   `json:"LastSeen"` // RFC3339
}

// CLISource invokes `tailscale status --json` to enumerate devices.
// If the binary is not found or the socket is not present, Devices returns
// an error immediately.
type CLISource struct{}

// Devices runs `tailscale status --json` with a 5-second timeout.
func (c CLISource) Devices(ctx context.Context) ([]Device, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tailscale", "status", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tailscale cli: exec failed: %w", err)
	}

	var status cliStatus
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("tailscale cli: parse JSON: %w", err)
	}

	devices := make([]Device, 0, 1+len(status.Peers))
	devices = append(devices, peerToDevice(status.Self))
	for _, peer := range status.Peers {
		devices = append(devices, peerToDevice(peer))
	}
	return devices, nil
}

func peerToDevice(p cliPeer) Device {
	d := Device{
		Hostname: p.HostName,
		Online:   p.Online,
	}
	if len(p.TailscaleIPs) > 0 {
		d.TailscaleIP = p.TailscaleIPs[0]
	}
	if p.LastSeen != "" {
		if t, err := time.Parse(time.RFC3339, p.LastSeen); err == nil {
			d.LastSeen = t
		}
	}
	return d
}
