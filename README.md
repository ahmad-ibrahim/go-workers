# go-workers
Simple & basic background workers for golang.

## How to use

Create a new dispatcher and specify the desired number of workers
```
dispatcher := workers.NewDispatcher(10)
```

Now register handlers for different types of jobs you need, you can assign multiple handlers for one job type.
```
dispatcher.RegisterHandler(
    // job category
    "billing.collect",
    // job handler
    func(ctx context.Context, payload string) error {
        //Do some work
        return nil
    },
    // an optional done handler that runs after the job is finished successfully
    func(ctx context.Context, key string) {
        // some after work, maybe acknowledgment using the job key
    },
    // an option failed handler that runs after the job failed
    func(ctx context.Context, err error, key string) {
        // some damage control like returning the job back to queue.
    },
    // Number of retries, 0 or more 
    0
)
```

Starting the dispatcher
```
dispatcher.Start()
```

Feeding jobs to the dispatcher
```
dispatcher.Enqueue(
    //Job category
    "billing.collect", 
    //Payload
    "10001",
    // Optional unique key you give to the job that can be used by the done or failed handlers
    "invoice#10001"
    )
```

Stopping the dispatcher when needed

```
dispatcher.Stop()
```