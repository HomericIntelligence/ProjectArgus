// Package nats provides JetStream durable subscriber infrastructure for the
// Atlas dashboard.  It connects to a NATS server, creates a durable push
// consumer for each configured stream, and publishes decoded Event values onto
// an EventBus for the rest of the application to consume.
package nats

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	natsgo "github.com/nats-io/nats.go"
)

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

// Config holds the connection and subscription parameters for the Subscriber.
type Config struct {
	// NATSURL is the NATS server URL, e.g. "nats://127.0.0.1:4222".
	NATSURL string
	// Streams is the list of JetStream stream configurations to subscribe to.
	Streams []StreamConfig
}

// StreamConfig describes a single JetStream stream and the durable consumer
// that Atlas will attach to it.
type StreamConfig struct {
	// Stream is the JetStream stream name, e.g. "homeric-agents".
	Stream string
	// Subjects is the list of NATS subjects to filter on, e.g. ["hi.agents.>"].
	Subjects []string
	// Durable is the durable consumer name, e.g. "atlas-agents".
	Durable string
}

// Event is the normalised representation of a NATS message published onto the
// EventBus.
type Event struct {
	// Topic is the high-level topic derived from the NATS subject, e.g. "agent".
	Topic string
	// Subject is the raw NATS subject, e.g. "hi.agents.host.name.heartbeat".
	Subject string
	// Payload is the raw JSON body of the message.
	Payload json.RawMessage
	// At is the time at which the event was received.
	At time.Time
}

// EventBus is the interface that the Subscriber uses to publish decoded events.
// Implementations must be safe for concurrent use from multiple goroutines.
type EventBus interface {
	Publish(e Event)
}

// Subscriber connects to NATS and maintains durable push consumers for each
// configured stream.
type Subscriber struct {
	cfg Config
	bus EventBus
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// New creates a Subscriber that will use cfg and bus.  It does not connect to
// NATS — connection is deferred to Start.
func New(cfg Config, bus EventBus) *Subscriber {
	return &Subscriber{cfg: cfg, bus: bus}
}

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------

// Start connects to the NATS server and subscribes to all configured streams.
// It blocks until ctx is cancelled, at which point it drains and closes the
// connection.  It returns a non-nil error if the initial connection fails.
func (s *Subscriber) Start(ctx context.Context) error {
	nc, err := natsgo.Connect(s.cfg.NATSURL, natsgo.MaxReconnects(0))
	if err != nil {
		return err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return err
	}

	for _, sc := range s.cfg.Streams {
		sc := sc // capture for closure
		handler := s.makeHandler(sc)
		_, err := js.Subscribe(
			sc.Subjects[0],
			handler,
			natsgo.Durable(sc.Durable),
			natsgo.DeliverNew(),
			natsgo.AckExplicit(),
			natsgo.AckWait(30*time.Second),
			natsgo.MaxAckPending(1024),
			natsgo.ConsumerFilterSubjects(sc.Subjects...),
		)
		if err != nil {
			// Best-effort: log and continue to other streams.
			_ = err
		}
	}

	// Block until context is cancelled.
	<-ctx.Done()

	// Drain and close the connection gracefully.
	_ = nc.Drain()
	return nil
}

// makeHandler returns a NATS MsgHandler that decodes the message and publishes
// an Event onto the bus.
func (s *Subscriber) makeHandler(sc StreamConfig) natsgo.MsgHandler {
	return func(msg *natsgo.Msg) {
		e := Event{
			Topic:   TopicFromSubject(msg.Subject),
			Subject: msg.Subject,
			Payload: json.RawMessage(msg.Data),
			At:      time.Now().UTC(),
		}
		_ = msg.Ack()
		s.bus.Publish(e)
	}
}

// ---------------------------------------------------------------------------
// DefaultStreams
// ---------------------------------------------------------------------------

// DefaultStreams returns the six canonical HomericIntelligence JetStream
// stream configurations that Atlas subscribes to.
func DefaultStreams() []StreamConfig {
	return []StreamConfig{
		{Stream: "homeric-agents", Subjects: []string{"hi.agents.>"}, Durable: "atlas-agents"},
		{Stream: "homeric-tasks", Subjects: []string{"hi.tasks.>"}, Durable: "atlas-tasks"},
		{Stream: "homeric-myrmidon", Subjects: []string{"hi.myrmidon.>"}, Durable: "atlas-myrmidon"},
		{Stream: "homeric-research", Subjects: []string{"hi.research.>"}, Durable: "atlas-research"},
		{Stream: "homeric-pipeline", Subjects: []string{"hi.pipeline.>"}, Durable: "atlas-pipeline"},
		{Stream: "homeric-logs", Subjects: []string{"hi.logs.>"}, Durable: "atlas-logs"},
	}
}

// ---------------------------------------------------------------------------
// TopicFromSubject
// ---------------------------------------------------------------------------

// TopicFromSubject derives a short topic label from a NATS subject.
// The mapping is:
//
//	hi.agents.*   → "agent"
//	hi.tasks.*    → "task"
//	hi.myrmidon.* → "myrmidon"
//	hi.research.* → "research"
//	hi.pipeline.* → "pipeline"
//	hi.logs.*     → "log"
//	anything else → "unknown"
func TopicFromSubject(subject string) string {
	parts := strings.SplitN(subject, ".", 3)
	if len(parts) < 2 || parts[0] != "hi" {
		return "unknown"
	}
	switch parts[1] {
	case "agents":
		return "agent"
	case "tasks":
		return "task"
	case "myrmidon":
		return "myrmidon"
	case "research":
		return "research"
	case "pipeline":
		return "pipeline"
	case "logs":
		return "log"
	default:
		return "unknown"
	}
}
