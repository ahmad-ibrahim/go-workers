package workers

import (
	"errors"
	"sync"
)

//Handler ...
type Handler struct {
	retries      int
	handler      JobHandler
	doneHandler  DoneHandler
	errorHandler ErrorHandler
}

//Dispatcher ...
type Dispatcher struct {
	handlers map[string][]*Handler
	job      chan job
	queue    chan chan job
	workers  []worker
	close    chan struct{}
	mx       sync.RWMutex
}

//NewDispatcher ...
func NewDispatcher(nWorkers int) (*Dispatcher, error) {
	if nWorkers < 1 {
		return nil, errors.New("workers count cannot be less than one")
	}
	d := &Dispatcher{
		handlers: make(map[string][]*Handler),
		queue:    make(chan chan job, nWorkers),
		job:      make(chan job, nWorkers),
		close:    make(chan struct{}),
	}

	for i := 1; i < nWorkers+1; i++ {
		worker := newWorker(i, d.queue)
		worker.start()
		d.workers = append(d.workers, worker)
	}

	return d, nil
}

//RegisterHandler ...
func (d *Dispatcher) RegisterHandler(topic string, handler JobHandler, doneHandler DoneHandler, errorHandler ErrorHandler, retries int) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	if retries < 0 {
		return errors.New("retries cannot bet less than 0")
	}

	h := &Handler{
		handler:      handler,
		retries:      retries,
		doneHandler:  doneHandler,
		errorHandler: errorHandler,
	}

	d.mx.Lock()
	d.handlers[topic] = append(d.handlers[topic], h)
	d.mx.Unlock()
	return nil
}

//Enqueue ...
func (d *Dispatcher) Enqueue(topic, payload, key string) error {
	if hs, ok := d.handlers[topic]; ok {
		for _, handler := range hs {
			h := handler
			go func() {
				d.job <- job{
					key:          key,
					payload:      payload,
					handler:      h.handler,
					retries:      h.retries,
					doneHandler:  h.doneHandler,
					errorHandler: h.errorHandler,
				}
			}()
		}
		return nil
	}

	return errors.New("unrecognized topic")
}

//Start ...
func (d *Dispatcher) Start() {
	go func() {
		for {
			select {
			case j := <-d.job:
				go func(j job) {
					worker := <-d.queue
					worker <- j
				}(j)
			case <-d.close:
				close(d.close)
				wg := &sync.WaitGroup{}
				for _, w := range d.workers {
					wg.Add(1)
					worker := w
					go func() {
						worker.stop()
						wg.Done()
					}()
				}
				wg.Wait()
				return
			}
		}
	}()
}

//Stop ...
func (d *Dispatcher) Stop() {
	go func() {
		d.close <- struct{}{}
	}()
}
