package workers

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	t.Run("Test handling payload", func(t *testing.T) {
		ch := make(chan string)
		w := &worker{
			key:     "1",
			payload: "Test",
			handler: func(ctx context.Context, payload string) error {
				ch <- payload
				return nil
			},
		}

		go func() { w.handle(context.Background()) }()

		result := <-ch
		if result != "Test" {
			t.Errorf("Wanted Test and got %s", result)
		}
	})

	t.Run("Test handling retires", func(t *testing.T) {
		ch := make(chan int)
		w := &worker{
			key:     "1",
			payload: "Test",
			retries: 3,
			handler: func(ctx context.Context, payload string) error {
				ch <- 1
				return errors.New("test error")
			},
		}

		go func() { w.handle(context.Background()) }()

		tries := 0

		for {
			select {
			case <-ch:
				tries++
				if tries == 3 {
					return
				}
			case <-time.After(time.Second * 2):
				if tries != 3 {
					t.Errorf("Wanted 3 and got %d", tries)
				}
				return
			}
		}
	})
}

func TestDone(t *testing.T) {
	ch := make(chan bool)
	w := &worker{
		key: "Test",
		handler: func(ctx context.Context, payload string) error {
			return nil
		},
		done: func(ctx context.Context, id string) {
			ch <- true
		},
	}

	go func() { w.handle(context.Background()) }()

	result := <-ch

	if !result {
		t.Error("The Done handler didn't return true to the test channel")
	}
}

func TestFailed(t *testing.T) {
	ch := make(chan error)
	testError := errors.New("Failed process")

	w := &worker{
		key: "Test",
		handler: func(ctx context.Context, payload string) error {
			return testError
		},
		failed: func(ctx context.Context, err error, id string) {
			ch <- err
		},
	}

	go func() { w.handle(context.Background()) }()

	result := <-ch

	if result != testError {
		t.Errorf("Wanted %s, got %s", testError, result)
	}
}
