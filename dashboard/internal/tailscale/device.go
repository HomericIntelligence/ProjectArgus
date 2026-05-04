package tailscale

import (
	"context"
	"time"
)

// Device represents a single node in the Tailscale mesh.
type Device struct {
	Hostname    string
	TailscaleIP string
	Online      bool
	LastSeen    time.Time
}

// Source is the interface satisfied by all Tailscale device sources.
// Implementations must be safe for concurrent use.
type Source interface {
	Devices(ctx context.Context) ([]Device, error)
}
