package server

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/HomericIntelligence/atlas/internal/version"
)

// pollSources are the pre-registered poller source labels for histograms.
var pollSources = []string{"agamemnon", "nestor", "hermes", "nats"}

// histogramBuckets are the upper bounds for atlas_poll_duration_seconds.
var histogramBuckets = []float64{0.1, 0.5, 1, 2, 5}

// numHistogramBuckets is the compile-time constant length of histogramBuckets.
const numHistogramBuckets = 5

// AtlasMetrics holds all Prometheus-compatible metrics for the Atlas dashboard.
// All counter fields use sync/atomic for lock-free increments. Labeled metrics
// use an RWMutex-protected map.
type AtlasMetrics struct {
	// atlas_nats_connected gauge (0 or 1)
	natsConnected atomic.Int64

	// atlas_sse_connected_clients gauge
	sseConnectedClients atomic.Int64

	// atlas_poll_errors_total{source} counter
	pollErrorsMu sync.RWMutex
	pollErrors   map[string]*atomic.Int64

	// atlas_sse_dropped_total{subscriber} counter
	sseDroppedMu sync.RWMutex
	sseDropped   map[string]*atomic.Int64

	// atlas_event_parse_errors_total{stream} counter
	eventParseErrorsMu sync.RWMutex
	eventParseErrors   map[string]*atomic.Int64

	// atlas_nats_messages_processed_total{stream} counter
	natsMessagesMu sync.RWMutex
	natsMessages   map[string]*atomic.Int64

	// atlas_poll_duration_seconds histogram, keyed by source
	histMu   sync.RWMutex
	histData map[string]*histogramData
}

// histogramData holds the bucket counts, sum, and total count for one histogram series.
type histogramData struct {
	buckets [numHistogramBuckets]atomic.Int64
	inf     atomic.Int64
	sum     atomic.Int64 // stored as nanoseconds; converted to seconds on output
	count   atomic.Int64
}

// newAtlasMetrics initialises an AtlasMetrics with pre-registered sources.
func newAtlasMetrics() *AtlasMetrics {
	m := &AtlasMetrics{
		pollErrors:       make(map[string]*atomic.Int64),
		sseDropped:       make(map[string]*atomic.Int64),
		eventParseErrors: make(map[string]*atomic.Int64),
		natsMessages:     make(map[string]*atomic.Int64),
		histData:         make(map[string]*histogramData),
	}
	// Pre-register all known poller sources so they always appear in output.
	for _, src := range pollSources {
		m.pollErrors[src] = &atomic.Int64{}
		m.histData[src] = &histogramData{}
	}
	return m
}

// --- Inc/Set helpers ---

// SetNATSConnected sets the atlas_nats_connected gauge (1 = connected, 0 = disconnected).
func (m *AtlasMetrics) SetNATSConnected(connected bool) {
	if connected {
		m.natsConnected.Store(1)
	} else {
		m.natsConnected.Store(0)
	}
}

// SetSSEConnectedClients updates the atlas_sse_connected_clients gauge.
func (m *AtlasMetrics) SetSSEConnectedClients(n int64) {
	m.sseConnectedClients.Store(n)
}

// IncPollError increments atlas_poll_errors_total for the given source label.
func (m *AtlasMetrics) IncPollError(source string) {
	m.pollErrorsMu.Lock()
	ctr, ok := m.pollErrors[source]
	if !ok {
		ctr = &atomic.Int64{}
		m.pollErrors[source] = ctr
	}
	m.pollErrorsMu.Unlock()
	ctr.Add(1)
}

// IncSSEDropped increments atlas_sse_dropped_total for the given subscriber label.
func (m *AtlasMetrics) IncSSEDropped(subscriber string) {
	m.sseDroppedMu.Lock()
	ctr, ok := m.sseDropped[subscriber]
	if !ok {
		ctr = &atomic.Int64{}
		m.sseDropped[subscriber] = ctr
	}
	m.sseDroppedMu.Unlock()
	ctr.Add(1)
}

// IncEventParseError increments atlas_event_parse_errors_total for the given stream label.
func (m *AtlasMetrics) IncEventParseError(stream string) {
	m.eventParseErrorsMu.Lock()
	ctr, ok := m.eventParseErrors[stream]
	if !ok {
		ctr = &atomic.Int64{}
		m.eventParseErrors[stream] = ctr
	}
	m.eventParseErrorsMu.Unlock()
	ctr.Add(1)
}

// IncNATSMessage increments atlas_nats_messages_processed_total for the given stream label.
func (m *AtlasMetrics) IncNATSMessage(stream string) {
	m.natsMessagesMu.Lock()
	ctr, ok := m.natsMessages[stream]
	if !ok {
		ctr = &atomic.Int64{}
		m.natsMessages[stream] = ctr
	}
	m.natsMessagesMu.Unlock()
	ctr.Add(1)
}

// ObservePollDuration records a poll duration (in seconds) for atlas_poll_duration_seconds.
func (m *AtlasMetrics) ObservePollDuration(source string, seconds float64) {
	m.histMu.Lock()
	hd, ok := m.histData[source]
	if !ok {
		hd = &histogramData{}
		m.histData[source] = hd
	}
	m.histMu.Unlock()

	// Increment all buckets where le >= seconds.
	for i, le := range histogramBuckets {
		if seconds <= le {
			hd.buckets[i].Add(1)
		}
	}
	hd.inf.Add(1)
	// Store sum as nanoseconds to avoid floating-point atomics.
	hd.sum.Add(int64(seconds * 1e9))
	hd.count.Add(1)
}

// --- Handler ---

// Handler returns an http.HandlerFunc that writes metrics in Prometheus text format.
func (m *AtlasMetrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		var b strings.Builder

		// atlas_build_info
		writeLine(&b, "# HELP atlas_build_info Build information about the Atlas dashboard.")
		writeLine(&b, "# TYPE atlas_build_info gauge")
		fmt.Fprintf(&b, "atlas_build_info{version=%q,goversion=%q} 1\n",
			version.Version, runtime.Version())

		// atlas_nats_connected
		writeLine(&b, "# HELP atlas_nats_connected 1 if the NATS connection is healthy, 0 otherwise.")
		writeLine(&b, "# TYPE atlas_nats_connected gauge")
		fmt.Fprintf(&b, "atlas_nats_connected %d\n", m.natsConnected.Load())

		// atlas_sse_connected_clients
		writeLine(&b, "# HELP atlas_sse_connected_clients Number of currently connected SSE clients.")
		writeLine(&b, "# TYPE atlas_sse_connected_clients gauge")
		fmt.Fprintf(&b, "atlas_sse_connected_clients %d\n", m.sseConnectedClients.Load())

		// atlas_poll_errors_total
		writeLine(&b, "# HELP atlas_poll_errors_total Total number of poller errors by source.")
		writeLine(&b, "# TYPE atlas_poll_errors_total counter")
		m.pollErrorsMu.RLock()
		for _, src := range sortedKeys(m.pollErrors) {
			fmt.Fprintf(&b, "atlas_poll_errors_total{source=%q} %d\n", src, m.pollErrors[src].Load())
		}
		m.pollErrorsMu.RUnlock()

		// atlas_sse_dropped_total
		writeLine(&b, "# HELP atlas_sse_dropped_total Total number of SSE events dropped for slow subscribers.")
		writeLine(&b, "# TYPE atlas_sse_dropped_total counter")
		m.sseDroppedMu.RLock()
		for _, sub := range sortedKeys(m.sseDropped) {
			fmt.Fprintf(&b, "atlas_sse_dropped_total{subscriber=%q} %d\n", sub, m.sseDropped[sub].Load())
		}
		m.sseDroppedMu.RUnlock()

		// atlas_event_parse_errors_total
		writeLine(&b, "# HELP atlas_event_parse_errors_total Total number of event parse errors by stream.")
		writeLine(&b, "# TYPE atlas_event_parse_errors_total counter")
		m.eventParseErrorsMu.RLock()
		for _, stream := range sortedKeys(m.eventParseErrors) {
			fmt.Fprintf(&b, "atlas_event_parse_errors_total{stream=%q} %d\n", stream, m.eventParseErrors[stream].Load())
		}
		m.eventParseErrorsMu.RUnlock()

		// atlas_nats_messages_processed_total
		writeLine(&b, "# HELP atlas_nats_messages_processed_total Total number of NATS messages processed by stream.")
		writeLine(&b, "# TYPE atlas_nats_messages_processed_total counter")
		m.natsMessagesMu.RLock()
		for _, stream := range sortedKeys(m.natsMessages) {
			fmt.Fprintf(&b, "atlas_nats_messages_processed_total{stream=%q} %d\n", stream, m.natsMessages[stream].Load())
		}
		m.natsMessagesMu.RUnlock()

		// atlas_poll_duration_seconds histogram
		writeLine(&b, "# HELP atlas_poll_duration_seconds Duration of poll requests in seconds.")
		writeLine(&b, "# TYPE atlas_poll_duration_seconds histogram")
		m.histMu.RLock()
		for _, src := range sortedHistKeys(m.histData) {
			hd := m.histData[src]
			for i, le := range histogramBuckets {
				fmt.Fprintf(&b, "atlas_poll_duration_seconds_bucket{source=%q,le=\"%g\"} %d\n",
					src, le, hd.buckets[i].Load())
			}
			fmt.Fprintf(&b, "atlas_poll_duration_seconds_bucket{source=%q,le=\"+Inf\"} %d\n",
				src, hd.inf.Load())
			// Convert nanoseconds back to seconds for the sum.
			sumSec := float64(hd.sum.Load()) / 1e9
			fmt.Fprintf(&b, "atlas_poll_duration_seconds_sum{source=%q} %g\n", src, sumSec)
			fmt.Fprintf(&b, "atlas_poll_duration_seconds_count{source=%q} %d\n", src, hd.count.Load())
		}
		m.histMu.RUnlock()

		_, _ = w.Write([]byte(b.String()))
	}
}

// writeLine writes a single line followed by a newline to b.
func writeLine(b *strings.Builder, s string) {
	b.WriteString(s)
	b.WriteByte('\n')
}

// sortedKeys returns the keys of a map[string]*atomic.Int64 in sorted order.
func sortedKeys(m map[string]*atomic.Int64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

// sortedHistKeys returns the keys of a map[string]*histogramData in sorted order.
func sortedHistKeys(m map[string]*histogramData) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

// sortStrings sorts a string slice in place (insertion sort — small N expected).
func sortStrings(ss []string) {
	for i := 1; i < len(ss); i++ {
		key := ss[i]
		j := i - 1
		for j >= 0 && ss[j] > key {
			ss[j+1] = ss[j]
			j--
		}
		ss[j+1] = key
	}
}
