package workers

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewDispatcher(t *testing.T) {
	t.Run("Test creating new dispatcher without error", func(t *testing.T) {
		_, err := NewDispatcher(10)
		if err != nil {
			t.Errorf("Expected no errors but got one: %s", err)
		}
	})

	t.Run("Test creating new dispatcher with error", func(t *testing.T) {
		_, err := NewDispatcher(0)
		if err == nil {
			t.Error("Expected error but got none.")
		}
	})
}

func TestRegisterHandler(t *testing.T) {
	d, _ := NewDispatcher(1)

	t.Run("Test registering a handler without error", func(t *testing.T) {
		err := d.RegisterHandler("category",
			func(ctx context.Context, payload string) error {
				return nil
			}, nil, nil, 1)

		if err != nil {
			t.Errorf("Expected no errors but got one: %s", err)
		}
	})

	t.Run("Test registering a handler with invalid handler", func(t *testing.T) {
		err := d.RegisterHandler("category", nil, nil, nil, 1)

		if err == nil {
			t.Error("Expected error but got none")
		}
	})
	t.Run("Test registering a handler with invalid retries", func(t *testing.T) {
		err := d.RegisterHandler("category",
			func(ctx context.Context, payload string) error {
				return nil
			}, nil, nil, -1)

		if err == nil {
			t.Error("Expected error but got none")
		}
	})

	t.Run("Test registering a handler and checking handlers count", func(t *testing.T) {
		d.RegisterHandler("category1",
			func(ctx context.Context, payload string) error {
				return nil
			}, nil, nil, 3)

		d.RegisterHandler("category1",
			func(ctx context.Context, payload string) error {
				return nil
			}, nil, nil, 3)

		d.RegisterHandler("category2",
			func(ctx context.Context, payload string) error {
				return nil
			}, nil, nil, 3)

		handlers := len(d.handlers["category1"]) + len(d.handlers["category2"])

		if handlers != 3 {
			t.Errorf("Expected 3 handlers but got %d", handlers)
		}
	})
}

func TestEnqueue(t *testing.T) {
	d, _ := NewDispatcher(1)
	d.RegisterHandler("topic",
		func(ctx context.Context, payload string) error {
			return nil
		}, nil, nil, 1)

	t.Run("Test adding work without error", func(t *testing.T) {
		err := d.Enqueue("topic", "test payload", "test key")
		if err != nil {
			t.Errorf("Expected no errors but got one: %s", err)
		}
	})

	t.Run("Test adding work with error", func(t *testing.T) {
		err := d.Enqueue("not topic", "test payload", "test key")
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestWorkDispatched(t *testing.T) {
	//ch := make(chan struct{})
	var counter int32
	wg := &sync.WaitGroup{}
	d, _ := NewDispatcher(10)
	d.RegisterHandler("topic",
		func(ctx context.Context, payload string) error {
			defer wg.Done()
			atomic.AddInt32(&counter, 1)
			return nil
		}, nil, nil, 1)

	d.Start()
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		d.Enqueue("topic", fmt.Sprintf("%d", i), fmt.Sprintf("%d", i))
	}

	wg.Wait()
	d.Stop()
	if counter != 100 {
		t.Errorf("Wanted 100 got %d", counter)
	}

}
