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
		w.queue <- w.job

		for {
			select {
			case job := <-w.job:
				ctx := context.Background()
				job.handle(ctx)
				w.queue <- w.job
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
