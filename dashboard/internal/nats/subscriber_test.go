package nats_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	atlnats "github.com/HomericIntelligence/atlas/internal/nats"
)

// ---------------------------------------------------------------------------
// mockBus — test double that satisfies EventBus
// ---------------------------------------------------------------------------

type mockBus struct {
	mu     sync.Mutex
	events []atlnats.Event
}

func (m *mockBus) Publish(e atlnats.Event) {
	m.mu.Lock()
	m.events = append(m.events, e)
	m.mu.Unlock()
}

func (m *mockBus) received() []atlnats.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]atlnats.Event(nil), m.events...)
}

// ---------------------------------------------------------------------------
// Compile-time interface check
// ---------------------------------------------------------------------------

// TestEventBusInterface verifies that *mockBus satisfies the EventBus interface.
// This test will fail to compile if the interface is not satisfied correctly.
func TestEventBusInterface(t *testing.T) {
	t.Parallel()
	var _ atlnats.EventBus = (*mockBus)(nil)
}

// ---------------------------------------------------------------------------
// canonical stream set
// ---------------------------------------------------------------------------

var canonicalStreams = []string{
	"homeric-agents",
	"homeric-tasks",
	"homeric-myrmidon",
	"homeric-research",
	"homeric-pipeline",
	"homeric-logs",
}

// ---------------------------------------------------------------------------
// TestDefaultStreams
// ---------------------------------------------------------------------------

// TestDefaultStreams asserts that DefaultStreams returns exactly 6 entries,
// each with non-empty fields, and that the stream names match the canonical list.
func TestDefaultStreams(t *testing.T) {
	t.Parallel()

	streams := atlnats.DefaultStreams()

	if len(streams) != 6 {
		t.Fatalf("DefaultStreams: want 6 entries, got %d", len(streams))
	}

	// Build a set of canonical names for O(1) lookup.
	canonicalSet := make(map[string]struct{}, len(canonicalStreams))
	for _, name := range canonicalStreams {
		canonicalSet[name] = struct{}{}
	}

	for i, sc := range streams {
		if sc.Stream == "" {
			t.Errorf("entry %d: Stream is empty", i)
		}
		if sc.Durable == "" {
			t.Errorf("entry %d: Durable is empty", i)
		}
		if len(sc.Subjects) == 0 {
			t.Errorf("entry %d: Subjects is empty", i)
		}
		for j, subj := range sc.Subjects {
			if subj == "" {
				t.Errorf("entry %d: Subjects[%d] is empty", i, j)
			}
		}

		if _, ok := canonicalSet[sc.Stream]; !ok {
			t.Errorf("entry %d: stream name %q is not in the canonical list", i, sc.Stream)
		}
	}

	// Verify every canonical stream name appears exactly once.
	seen := make(map[string]int, 6)
	for _, sc := range streams {
		seen[sc.Stream]++
	}
	for _, name := range canonicalStreams {
		if seen[name] != 1 {
			t.Errorf("canonical stream %q appears %d times, want 1", name, seen[name])
		}
	}
}

// ---------------------------------------------------------------------------
// TestDefaultStreamsDurablePrefix
// ---------------------------------------------------------------------------

// TestDefaultStreamsDurablePrefix asserts that every durable consumer name
// in DefaultStreams starts with the "atlas-" prefix.
func TestDefaultStreamsDurablePrefix(t *testing.T) {
	t.Parallel()

	const prefix = "atlas-"
	for _, sc := range atlnats.DefaultStreams() {
		if !strings.HasPrefix(sc.Durable, prefix) {
			t.Errorf("stream %q: durable name %q does not start with %q",
				sc.Stream, sc.Durable, prefix)
		}
	}
}

// ---------------------------------------------------------------------------
// TestNewSubscriberNoConnect
// ---------------------------------------------------------------------------

// TestNewSubscriberNoConnect verifies that New returns a non-nil Subscriber
// without panicking and without dialling NATS. We prove the latter by
// supplying an invalid URL — if New dialled it would either panic or return
// an error (it must not).
func TestNewSubscriberNoConnect(t *testing.T) {
	t.Parallel()

	cfg := atlnats.Config{
		NATSURL: "nats://127.0.0.1:1", // port 1 is never open
		Streams: atlnats.DefaultStreams(),
	}
	bus := &mockBus{}

	// Must not panic.
	sub := atlnats.New(cfg, bus)
	if sub == nil {
		t.Fatal("New returned nil; want non-nil *Subscriber")
	}

	// No events should have been published during construction.
	if got := bus.received(); len(got) != 0 {
		t.Errorf("New published %d events during construction, want 0", len(got))
	}
}

// ---------------------------------------------------------------------------
// TestSubscriberStartConnectFailure
// ---------------------------------------------------------------------------

// TestSubscriberStartConnectFailure asserts that Start returns a non-nil error
// when NATS is unreachable. The test enforces a 3-second deadline so that a
// hung connection attempt does not block the suite indefinitely.
func TestSubscriberStartConnectFailure(t *testing.T) {
	t.Parallel()

	cfg := atlnats.Config{
		NATSURL: "nats://127.0.0.1:1", // guaranteed unreachable
		Streams: atlnats.DefaultStreams(),
	}
	bus := &mockBus{}
	sub := atlnats.New(cfg, bus)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := sub.Start(ctx)
	if err == nil {
		t.Fatal("Start: want non-nil error for unreachable NATS server, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestEventTopicFromSubject
// ---------------------------------------------------------------------------

// TestEventTopicFromSubject verifies the TopicFromSubject helper using the
// canonical subject-to-topic mapping required by the package design.
func TestEventTopicFromSubject(t *testing.T) {
	t.Parallel()

	cases := []struct {
		subject string
		want    string
	}{
		{"hi.agents.host.name.heartbeat", "agent"},
		{"hi.tasks.team.task.completed", "task"},
		{"hi.myrmidon.host.name.result", "myrmidon"},
		{"unknown.subject", "unknown"},
	}

	for _, tc := range cases {
		tc := tc // capture loop variable
		t.Run(tc.subject, func(t *testing.T) {
			t.Parallel()
			got := atlnats.TopicFromSubject(tc.subject)
			if got != tc.want {
				t.Errorf("TopicFromSubject(%q) = %q, want %q", tc.subject, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestEventStruct — sanity-check the Event type is usable
// ---------------------------------------------------------------------------

// TestEventStruct ensures the Event struct fields are accessible and that
// json.RawMessage Payload round-trips correctly.
func TestEventStruct(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{"key":"value"}`)
	e := atlnats.Event{
		Topic:   "agent",
		Subject: "hi.agents.host.name.heartbeat",
		Payload: raw,
		At:      time.Now().UTC(),
	}

	if e.Topic == "" {
		t.Error("Event.Topic is empty")
	}
	if e.Subject == "" {
		t.Error("Event.Subject is empty")
	}
	if len(e.Payload) == 0 {
		t.Error("Event.Payload is empty")
	}
	if e.At.IsZero() {
		t.Error("Event.At is zero")
	}
}

// ---------------------------------------------------------------------------
// TestConfigFields — ensure Config and StreamConfig are addressable
// ---------------------------------------------------------------------------

// TestConfigFields verifies that Config and StreamConfig can be constructed
// with all required fields, preventing silent zero-value bugs.
func TestConfigFields(t *testing.T) {
	t.Parallel()

	sc := atlnats.StreamConfig{
		Stream:   "homeric-agents",
		Subjects: []string{"hi.agents.>"},
		Durable:  "atlas-agents",
	}
	cfg := atlnats.Config{
		NATSURL: "nats://127.0.0.1:4222",
		Streams: []atlnats.StreamConfig{sc},
	}

	if cfg.NATSURL == "" {
		t.Error("Config.NATSURL is empty")
	}
	if len(cfg.Streams) != 1 {
		t.Fatalf("Config.Streams: want 1, got %d", len(cfg.Streams))
	}
	if cfg.Streams[0].Stream != "homeric-agents" {
		t.Errorf("StreamConfig.Stream: want %q, got %q", "homeric-agents", cfg.Streams[0].Stream)
	}
	if cfg.Streams[0].Durable != "atlas-agents" {
		t.Errorf("StreamConfig.Durable: want %q, got %q", "atlas-agents", cfg.Streams[0].Durable)
	}
}
