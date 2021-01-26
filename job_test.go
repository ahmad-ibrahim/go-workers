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
		j := &job{
			key:     "1",
			payload: "Test",
			handler: func(ctx context.Context, payload string) error {
				ch <- payload
				return nil
			},
		}

		go func() { j.handle(context.Background()) }()

		result := <-ch
		if result != "Test" {
			t.Errorf("Wanted Test and got %s", result)
		}
	})

	t.Run("Test handling retires", func(t *testing.T) {
		ch := make(chan int)
		j := &job{
			key:     "1",
			payload: "Test",
			retries: 3,
			handler: func(ctx context.Context, payload string) error {
				ch <- 1
				return errors.New("test error")
			},
		}

		go func() { j.handle(context.Background()) }()

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

	t.Run("Test handling panics", func(t *testing.T) {
		j := &job{
			key:     "1",
			payload: "Test",
			handler: func(ctx context.Context, payload string) error {
				panic("test error")
			},
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Not handled panic: %s", r)
			}
		}()
		j.handle(context.Background())
	})

}

func TestDone(t *testing.T) {
	ch := make(chan bool)
	j := &job{
		key: "Test",
		handler: func(ctx context.Context, payload string) error {
			return nil
		},
		doneHandler: func(ctx context.Context, id string) {
			ch <- true
		},
	}

	go func() { j.handle(context.Background()) }()

	result := <-ch

	if !result {
		t.Error("The Done handler didn't return true to the test channel")
	}
}

func TestFailed(t *testing.T) {
	ch := make(chan error)
	testError := errors.New("Failed process")

	j := &job{
		key: "Test",
		handler: func(ctx context.Context, payload string) error {
			return testError
		},
		errorHandler: func(ctx context.Context, err error, id string) {
			ch <- err
		},
	}

	go func() { j.handle(context.Background()) }()

	result := <-ch

	if result != testError {
		t.Errorf("Wanted %s, got %s", testError, result)
	}
}
