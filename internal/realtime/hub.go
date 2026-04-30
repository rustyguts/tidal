package realtime

import (
	"sync"
	"time"
)

// EventKind enumerates the event types broadcast on a job topic.
type EventKind string

const (
	EventStatus   EventKind = "status"
	EventProgress EventKind = "progress"
	EventLog      EventKind = "log"
)

type Event struct {
	Kind  EventKind   `json:"kind"`
	Topic string      `json:"topic"`
	At    time.Time   `json:"at"`
	Data  any `json:"data"`
}

const TopicJobsFirehose = "jobs"

// Hub is an in-memory pub/sub keyed by topic string.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]map[*subscription]struct{}
	bufferSize  int
}

type subscription struct {
	ch chan Event
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[*subscription]struct{}),
		bufferSize:  64,
	}
}

// Subscribe returns a receive channel for the given topic plus a cancel func
// that removes the subscription and closes the channel.
func (h *Hub) Subscribe(topic string) (<-chan Event, func()) {
	sub := &subscription{ch: make(chan Event, h.bufferSize)}
	h.mu.Lock()
	subs, ok := h.subscribers[topic]
	if !ok {
		subs = make(map[*subscription]struct{})
		h.subscribers[topic] = subs
	}
	subs[sub] = struct{}{}
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		if subs, ok := h.subscribers[topic]; ok {
			if _, present := subs[sub]; present {
				delete(subs, sub)
				close(sub.ch)
			}
			if len(subs) == 0 {
				delete(h.subscribers, topic)
			}
		}
		h.mu.Unlock()
	}
	return sub.ch, cancel
}

// Publish dispatches the event to all subscribers of `topic` plus the firehose.
// Slow consumers drop the event rather than block.
func (h *Hub) Publish(topic string, kind EventKind, data any) {
	ev := Event{Kind: kind, Topic: topic, At: time.Now(), Data: data}
	h.fanout(topic, ev)
	if topic != TopicJobsFirehose {
		h.fanout(TopicJobsFirehose, ev)
	}
}

func (h *Hub) fanout(topic string, ev Event) {
	h.mu.RLock()
	subs := h.subscribers[topic]
	for sub := range subs {
		select {
		case sub.ch <- ev:
		default:
			// drop on backpressure
		}
	}
	h.mu.RUnlock()
}
