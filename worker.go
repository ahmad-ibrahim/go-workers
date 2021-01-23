package workers

import "context"

//Worker ...
type worker struct {
	key     string
	retries int
	payload string
	handler JobHandler
	done    DoneHandler
	failed  FailedHandler
}

//JobHandler ...
type JobHandler func(ctx context.Context, payload string) error

//DoneHandler ...
type DoneHandler func(ctx context.Context, key string)

//FailedHandler ...
type FailedHandler func(ctx context.Context, err error, key string)

func (w *worker) handle(ctx context.Context) {
	var err error

	for i := 0; i < w.retries+1; i++ {
		err = w.handler(ctx, w.payload)
		if err == nil {
			break
		}
	}

	if err != nil {
		if w.failed != nil {
			w.failed(ctx, err, w.key)
		}
	} else {
		if w.done != nil {
			w.done(ctx, w.key)
		}
	}
}
