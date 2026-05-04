package events

import (
	"sync"
	"sync/atomic"
)

// ringBuffer is a fixed-capacity circular buffer of Events.
// When full, Push overwrites the oldest entry.
type ringBuffer struct {
	buf  []Event
	cap  int
	head int // index of the oldest entry (valid when size > 0)
	tail int // index where the next entry will be written
	size int
}

func newRingBuffer(cap int) *ringBuffer {
	if cap <= 0 {
		cap = 1
	}
	return &ringBuffer{
		buf: make([]Event, cap),
		cap: cap,
	}
}

// Push appends e to the ring, overwriting the oldest entry when full.
func (r *ringBuffer) Push(e Event) {
	r.buf[r.tail] = e
	r.tail = (r.tail + 1) % r.cap
	if r.size < r.cap {
		r.size++
	} else {
		// Overwrite: advance head past the overwritten slot.
		r.head = (r.head + 1) % r.cap
	}
}

// Snapshot returns up to n most-recent events in chronological (oldest-first)
// order. The returned slice is a copy.
func (r *ringBuffer) Snapshot(n int) []Event {
	if r.size == 0 {
		return nil
	}
	count := r.size
	if n < count {
		count = n
	}
	out := make([]Event, count)
	// The ring contains r.size events; the oldest is at r.head.
	// We want the last `count` of them, so start at offset (r.size - count)
	// from the head.
	start := (r.head + (r.size - count)) % r.cap
	for i := range count {
		out[i] = r.buf[(start+i)%r.cap]
	}
	return out
}

// Bus is an in-process pub/sub event bus with a ring-buffer history and
// non-blocking fan-out to registered subscribers.
type Bus struct {
	mu   sync.RWMutex
	subs map[chan Event]struct{}
	ring *ringBuffer
	// drops counts events not delivered due to a full subscriber channel.
	drops int64
}

// NewBus returns a new Bus whose ring buffer holds at most ringCap events.
func NewBus(ringCap int) *Bus {
	return &Bus{
		subs: make(map[chan Event]struct{}),
		ring: newRingBuffer(ringCap),
	}
}

// Subscribe returns a new channel that will receive published events.
// bufSize is the channel's buffer depth; use 0 for an unbuffered channel.
func (b *Bus) Subscribe(bufSize int) <-chan Event {
	ch := make(chan Event, bufSize)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes ch from the subscriber set and closes it.
func (b *Bus) Unsubscribe(ch <-chan Event) {
	b.mu.Lock()
	// The map key type must be chan Event (bidirectional) so we need to
	// recover it. We iterate to find the matching channel.
	for k := range b.subs {
		if k == ch {
			delete(b.subs, k)
			close(k)
			break
		}
	}
	b.mu.Unlock()
}

// Publish records e in the ring buffer and attempts a non-blocking send to
// every registered subscriber. Events dropped due to a full channel increment
// the drop counter atomically.
func (b *Bus) Publish(e Event) {
	b.mu.Lock()
	b.ring.Push(e)
	// Snapshot the subscriber set while holding the write lock so we can
	// send without holding the lock (avoiding potential deadlock if a
	// subscriber's select blocks, though our send is non-blocking anyway).
	// We collect references and send inside the lock because the send itself
	// is non-blocking (select + default).
	for ch := range b.subs {
		select {
		case ch <- e:
		default:
			atomic.AddInt64(&b.drops, 1)
		}
	}
	b.mu.Unlock()
}

// Snapshot returns up to n of the most-recent events in chronological
// (oldest-first) order.
func (b *Bus) Snapshot(n int) []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.ring.Snapshot(n)
}

// Drops returns the total number of events that were dropped because a
// subscriber's channel buffer was full at the time of publish.
func (b *Bus) Drops() int64 {
	return atomic.LoadInt64(&b.drops)
}
