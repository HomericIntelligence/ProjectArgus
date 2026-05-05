package store

import (
	"encoding/json"
	"time"

	"github.com/HomericIntelligence/atlas/internal/catalog"
)

// NATSStreamInfo holds per-stream data scraped from the NATS /jsz?detail=1 endpoint.
type NATSStreamInfo struct {
	Name      string    `json:"name"`
	Subjects  []string  `json:"subjects"`
	Messages  uint64    `json:"num_msgs"`
	Bytes     uint64    `json:"bytes"`
	Consumers int       `json:"num_consumers"`
	Created   time.Time `json:"created"`
}

// NATSConsumerInfo holds per-consumer data for a JetStream stream.
type NATSConsumerInfo struct {
	Stream        string    `json:"stream_name"`
	Durable       string    `json:"durable_name"`
	AckPending    int       `json:"num_ack_pending"`
	NumPending    uint64    `json:"num_pending"`
	LastDelivered time.Time `json:"ts"`
}

// NATSConnInfo holds per-connection data scraped from the NATS /connz endpoint.
type NATSConnInfo struct {
	Name          string `json:"name"`
	IP            string `json:"ip"`
	Subscriptions int    `json:"subscriptions"`
	InMsgs        int64  `json:"in_msgs"`
	OutMsgs       int64  `json:"out_msgs"`
	Uptime        string `json:"uptime"`
}

// HostView is the JSON representation of a host and its service probe results.
type HostView struct {
	Hostname    string                `json:"hostname"`
	TailscaleIP string                `json:"tailscale_ip"`
	Online      bool                  `json:"online"`
	Services    []catalog.ProbeResult `json:"services"`
}

// AgentRecord represents a single agent reported by Agamemnon.
type AgentRecord struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskRecord represents a single task reported by Agamemnon.
type TaskRecord struct {
	ID         string    `json:"id"`
	TeamID     string    `json:"team_id"`
	Subject    string    `json:"subject"`
	Status     string    `json:"status"`
	AssigneeID string    `json:"assignee_agent_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NATSStats holds key statistics scraped from the NATS monitoring endpoints.
type NATSStats struct {
	Connections int   `json:"connections"`
	Streams     int   `json:"streams"`
	InMsgs      int64 `json:"in_msgs"`
	OutMsgs     int64 `json:"out_msgs"`
}

// RawEvent is a single raw NATS event captured for an agent or task history tail.
type RawEvent struct {
	Topic      string          `json:"topic"`
	Subject    string          `json:"subject"`
	Payload    json.RawMessage `json:"payload"`
	ReceivedAt time.Time       `json:"received_at"`
}
