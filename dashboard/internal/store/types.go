package store

import "github.com/HomericIntelligence/atlas/internal/catalog"

// HostView is the JSON representation of a host and its service probe results.
type HostView struct {
	Hostname    string                 `json:"hostname"`
	TailscaleIP string                 `json:"tailscale_ip"`
	Online      bool                   `json:"online"`
	Services    []catalog.ProbeResult  `json:"services"`
}
