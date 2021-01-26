package workers

import (
	"context"
	"testing"
	"time"
)

func TestWorkerStart(t *testing.T) {
	queue := make(chan chan job, 1)
	w := newWorker(1, queue)
	w.start()
	worker := <-queue
	ch := make(chan struct{})
	worker <- job{
		key:     "Test key",
		payload: "Test payload",
		retries: 0,
		handler: func(ctx context.Context, payload string) error {
			ch <- struct{}{}
			return nil
		},
	}

	for {
		select {
		case <-ch:
			w.stop()
			return
		case <-time.After(time.Second * 2):
			w.stop()
			t.Error("Failed to get response from the handler")
			return
		}
	}
}

func TestWorkerStop(t *testing.T) {
	w := newWorker(1, make(chan chan job, 1))
	w.start()
	w.stop()
	time.Sleep(time.Second)
	if w.active {
		t.Error("Worker is still active.")
	}
}
