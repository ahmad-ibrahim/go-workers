package workers

import (
	"context"
	"errors"
)

//Dispatcher ...
type Dispatcher struct {
	handlers map[string][]*worker
	jobs     chan *worker
}

//NewDispatcher ...
func NewDispatcher(workersCount int) (*Dispatcher, error) {
	if workersCount < 1 {
		return nil, errors.New("workers count cannot be less than one")
	}
	return &Dispatcher{
		handlers: make(map[string][]*worker),
		jobs:     make(chan *worker, workersCount),
	}, nil
}

//RegisterHandler ...
func (d *Dispatcher) RegisterHandler(category string, handler JobHandler, done DoneHandler, failed FailedHandler, retries int) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	if retries < 0 {
		return errors.New("retries cannot bet less than 0")
	}

	w := &worker{
		retries: retries,
		handler: handler,
		done:    done,
		failed:  failed,
	}

	d.handlers[category] = append(d.handlers[category], w)
	return nil
}

//AddWork ...
func (d *Dispatcher) AddWork(category, payload, key string) error {
	if hs, ok := d.handlers[category]; ok {
		for _, handler := range hs {
			h := handler
			go func() {
				d.jobs <- &worker{
					key:     key,
					payload: payload,
					handler: h.handler,
					done:    h.done,
					failed:  h.failed,
					retries: h.retries,
				}
			}()
		}
		return nil
	}

	return errors.New("unrecognized category")
}

//Start ...
func (d *Dispatcher) Start() {
	go func() {
		for w := range d.jobs {
			worker := w
			go func() {
				worker.handle(context.Background())
			}()
		}
	}()
}

//Stop ...
func (d *Dispatcher) Close() {
	close(d.jobs)
}
