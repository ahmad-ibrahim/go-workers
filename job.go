package workers

import (
	"context"
	"log"
)

//JobHandler ...
type JobHandler func(ctx context.Context, payload string) error

//DoneHandler ...
type DoneHandler func(ctx context.Context, key string)

//ErrorHandler ...
type ErrorHandler func(ctx context.Context, err error, key string)

type job struct {
	key          string
	retries      int
	payload      string
	handler      JobHandler
	doneHandler  DoneHandler
	errorHandler ErrorHandler
}

func (j *job) panicHandler() {
	if r := recover(); r != nil {
		log.Printf("Job with key %s paniced.", r)
	}
}
func (j *job) handle(ctx context.Context) {
	defer j.panicHandler()
	var err error

	for i := 0; i < j.retries+1; i++ {
		err = j.handler(ctx, j.payload)
		if err == nil {
			break
		}
	}

	if err != nil {
		if j.errorHandler != nil {
			j.errorHandler(ctx, err, j.key)
		}
	} else {
		if j.doneHandler != nil {
			j.doneHandler(ctx, j.key)
		}
	}
}
