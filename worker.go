package workers

import "context"

type worker struct {
	id     int
	job    chan job
	queue  chan chan job
	active bool
	quit   chan struct{}
}

func newWorker(id int, queue chan chan job) worker {
	return worker{
		id:    id,
		job:   make(chan job),
		queue: queue,
		quit:  make(chan struct{}),
	}
}

func (w *worker) start() {
	w.active = true
	go func() {
		for {
			// Add worker to the queue
			w.queue <- w.job

			select {
			case job := <-w.job:
				ctx := context.Background()
				job.handle(ctx)
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *worker) stop() {
	w.active = false
	w.quit <- struct{}{}
}
