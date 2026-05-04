package events

import (
	"encoding/json"
	"time"
)

// Event is a single observable occurrence within the Atlas dashboard.
// Topic identifies the broad category (e.g. "agent", "task", "nats") and
// Subject is the fully-qualified NATS-style subject string.
type Event struct {
	Topic   string
	Subject string
	Payload json.RawMessage
	At      time.Time
}
