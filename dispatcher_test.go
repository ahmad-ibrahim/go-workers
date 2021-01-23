package workers

import (
	"context"
	"testing"
	"time"
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

	t.Run("Test registering a handler and checking workers count", func(t *testing.T) {
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

func TestAddWork(t *testing.T) {
	d, _ := NewDispatcher(1)
	d.RegisterHandler("category",
		func(ctx context.Context, payload string) error {
			return nil
		}, nil, nil, 1)

	t.Run("Test adding work without error", func(t *testing.T) {
		err := d.AddWork("category", "test payload", "test key")
		if err != nil {
			t.Errorf("Expected no errors but got one: %s", err)
		}
	})

	t.Run("Test adding work with error", func(t *testing.T) {
		err := d.AddWork("not category", "test payload", "test key")
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestWorkDispatched(t *testing.T) {
	ch := make(chan struct{})

	d, _ := NewDispatcher(1)
	d.RegisterHandler("category",
		func(ctx context.Context, payload string) error {
			ch <- struct{}{}
			return nil
		}, nil, nil, 1)

	d.AddWork("category", "test payload", "test key")
	d.Start()

	select {
	case <-ch:
		d.Stop()
		return
	case <-time.After(time.Second * 10):
		d.Stop()
		t.Error("Failed to get response from the dispatched job")
		return
	}
}
