package realtime

import (
	"sync"
	"testing"
	"time"
)

func TestHub_SubscribePublish(t *testing.T) {
	h := NewHub()

	ch, cancel := h.Subscribe("test-topic")
	defer cancel()

	h.Publish("test-topic", EventStatus, "hello")

	select {
	case ev := <-ch:
		if ev.Kind != EventStatus {
			t.Errorf("Kind = %q, want %q", ev.Kind, EventStatus)
		}
		if ev.Topic != "test-topic" {
			t.Errorf("Topic = %q", ev.Topic)
		}
		if got, ok := ev.Data.(string); !ok || got != "hello" {
			t.Errorf("Data = %v, want hello", ev.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestHub_Firehose(t *testing.T) {
	h := NewHub()

	firehoseCh, firehoseCancel := h.Subscribe(TopicJobsFirehose)
	defer firehoseCancel()

	jobCh, jobCancel := h.Subscribe("job:123")
	defer jobCancel()

	h.Publish("job:123", EventProgress, map[string]int{"percent": 50})

	select {
	case <-firehoseCh:
	case <-time.After(time.Second):
		t.Fatal("firehose should have received event")
	}

	select {
	case <-jobCh:
	case <-time.After(time.Second):
		t.Fatal("job subscriber should have received event")
	}
}

func TestHub_Unsubscribe(t *testing.T) {
	h := NewHub()

	ch, cancel := h.Subscribe("topic")
	cancel()

	h.Publish("topic", EventStatus, "data")

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed after cancel")
		}
	default:
	}
}

func TestHub_BackpressureDrops(t *testing.T) {
	h := NewHub()
	h.bufferSize = 1

	ch, cancel := h.Subscribe("topic")
	defer cancel()

	for i := 0; i < 5; i++ {
		h.Publish("topic", EventStatus, i)
	}

	select {
	case <-ch:
	default:
		t.Fatal("expected at least one event")
	}

	select {
	case <-ch:
		t.Error("expected no more events, buffer should have dropped")
	default:
	}
}

func TestHub_ConcurrentPublish(t *testing.T) {
	h := NewHub()

	_, cancel := h.Subscribe("topic")
	defer cancel()

	var wg sync.WaitGroup
	n := 50
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			h.Publish("topic", EventStatus, "data")
		}()
	}

	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()

	select {
	case <-c:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for concurrent publishes")
	}
}

func TestHub_PublishToOwnTopicOnly(t *testing.T) {
	h := NewHub()

	firehoseCh, firehoseCancel := h.Subscribe(TopicJobsFirehose)
	defer firehoseCancel()

	h.Publish(TopicJobsFirehose, EventStatus, "global")

	select {
	case <-firehoseCh:
	case <-time.After(time.Second):
		t.Fatal("firehose subscriber should get direct publish")
	}
}