# go-work
go-work is a framework that takes many of the http framework abstractions (requests, status, middleware, etc) and applies them to generic work loads.


## Why go-work?
It seems like there's always work to do in non-trivial systems.  Some times that work can easily be wrapped into an gRPC or HTTP service, but there are many problems that just don't fix that pattern:
- deferred background Jobs from an gRPC/HTTP request
- adding/consuming Jobs from a queue (Redis, Pulsar, S3, etc)
- adding/consuming Jobs from a database table (just another type of queue)
- scheduling Jobs in the future

Whenever we tackle these work problems, we need to solve the same issues for every execution of a Job:
- logging  
- metrics 
- tracing
- concurrency

If we do have concurrent work, then we also need to manage things like:
- on-ramping the load
- max concurrency
- cancelling Jobs


# Using go-work

## [Jobs](https://github.com/Bose/go-work/blob/master/docs.md#type-job)
Jobs define the work to be done in a common interface.  Think of them like an http.Request with an attached context (Ctx).   Jobs have Args, Context, and optional timeouts.   A Job.Ctx has key/values which can be used to pass data down the chain of adapters (aka middleware).  A Job.Ctx.Status is set by a Handler to represent the state of the Job's execution: Success, Error, NoResponse, etc)  Jobs are passed to Handlers by a Worker for each request.   
``` go
type Job struct {
    // Queue the job should be placed into
    Queue string

    // ctx related to the execution of a job - Perform(job) gets a new ctx everytime
    Ctx *Context

    // Args that will be passed to the Handler as the 2nd parameter when run
    Args Args

    // Handler that will be run by the worker
    Handler string

    // Timeout for every execution of job
    Timeout time.Duration
}

```
## [Handlers](https://github.com/Bose/go-work/blob/master/docs.md#type-handler)
Handlers define the func to be executed for a Job by a Worker.  Handlers also represent an interface than can be chained together to create middleware (aka Adapters)
``` go
type Handler func(j *Job) error
```

Example that "handles" a publishing Job request.  You can imagine how easy it will be to define a new handler that handles publishing the next message from a Redis source.   Notice the handler sets the Job's status before returning.
```go
func DefaultPublishNextMessageCRDB(j *work.Job) error {
	box := j.Args["box"].(outbox.EventOutbox)
	publisher := j.Args["publisher"].(pulsar.Producer)
	connFactory := j.Args["outboxConnectionFactory"].(OutboxConnectionFactory)

	l, ok := j.Ctx.Get(work.AggregateLogger)
	if !ok {
		err := fmt.Errorf("PublishNextMessage: no logger")
		logrus.Error(err)
		j.Ctx.SetStatus(work.StatusInternalError)
		return err
	}
	log := l.(*logrus.Logger)
	logger := log.WithFields(logrus.Fields{})

	db, err := connFactory(logger)
	if err != nil {
		j.Ctx.SetStatus(work.StatusInternalError)
		logger.Error(err)
		return err
	}
	logger.Infof("PublishNextMessage: checking new messages...")
	toPub, err := box.NextMessage(db, logger)
	if err != nil {
		logger.Infof("PublishNextMessage: no message: %s", err.Error())
		j.Ctx.SetStatus(work.StatusNoResponse)
		return nil
	}

	event := createEvent(toPub)
	if err := pubsub.PublishEvent(context.Background(), publisher, event); err != nil {
		j.Ctx.SetStatus(work.StatusInternalError)
		err := fmt.Errorf("PublishNextMessage: pubsub.PublishEvent error == %s", err.Error())
		logger.Error(err)
		return err
	}
	logger.Infof("PublishNextMessage: published to topic %s", publisher.Topic())

	if err := box.MarkAsPublished(db, toPub, logger, outbox.WithPublishedToTopics(publisher.Topic())); err != nil {
		j.Ctx.SetStatus(work.StatusInternalError)
		err := fmt.Errorf("PublishNextMessage: box.MarkAsPublished error == %s", err.Error())
		logger.Error(err)
		return err
	}
	j.Ctx.SetStatus(work.StatusSuccess)
	return nil
}
```

## [Concurrent Job](https://github.com/Bose/go-work/blob/master/docs.md#type-concurrentjob)
Concurrent jobs define a job to be performed by workers concurrently.

``` go
type ConcurrentJob struct {

    // PerformWith if the job will be performed with PerformWithEveryWithSync as a reoccuring job
    // or just once as PerformWithWithSync
    PerformWith PerformWith

    // PerformEvery defines the duration between executions of PerformWithEveryWithSync jobs
    PerformEvery time.Duration

    // MaxWorkers for the concurrent Job
    MaxWorkers int64

    // Job to run concurrently
    Job Job
    // contains filtered or unexported fields
}
```
### ConcurrentJob API
```
func (j *ConcurrentJob) Start() error
func (j *ConcurrentJob) Stop()
func (j *ConcurrentJob) Register(name string, h Handler) error
func (*ConcurrentJob) RunningWorkers
func NewConcurrentJob(
    job Job,
    workerFactory WorkerFactory,
    performWith PerformWith,
    performEvery time.Duration,
    maxWorker int64,
    startInterval time.Duration,
    logger Logger,
) (ConcurrentJob, error)
```
Be sure to call ConcurrentJob.Stop() or your program will leak resources.
```go
defer w.Stop()
```
### WorkerFactory API 
```
type WorkerFactory func(context.Context, Logger) Worker
func NewCommonWorkerFactory(ctx context.Context, l Logger) Worker

```

## Job.Ctx Status
Handlers set the status for every Job request (execution) using Job.Ctx.SetStatus() 
This allows middleware to take action based on the Job's status for things like metrics and logging.
``` go
const (
    // StatusUnknown was a job with an unknown status
    StatusUnknown Status = -1

    // StatusSucces was a successful job
    StatusSuccess = 200

    // StatusBadRequest was a job with a bad request
    StatusBadRequest = 400

    // StatusForbidden was a forbidden job
    StatusForbidden = 403

    // StatusUnauthorized was an unauthorized job
    StatusUnauthorized = 401

    // StatusTimeout was a job that timed out
    StatusTimeout = 408

    // StatusNoResponse was a job that intentionally created no response (basically the conditions were met for a noop by the Job)
    StatusNoResponse = 444

    // StatusInternalError was a job with an internal error
    StatusInternalError = 500

    // StatusUnavailable was a job that was unavailable
    StatusUnavailable = 503
)
```

## [Worker](https://github.com/Bose/go-work/blob/master/docs.md#type-worker)
A Worker implements an interface that defines how a Job will be executed.  
- now (sync and async)
- at a time in the future (only async)
- after waiting a specific time (only async)
- every occurence of a specified time interval (sync and async)

Be sure to call Worker.Stop() or your program could leak resources.
```go
defer w.Stop()
```

Official implementations:  
- [CommonWorker](https://github.com/Bose/go-work/blob/master/docs.md#type-worker). CommonWorkers is backed by the standard lib and goroutines. 

``` go
type Worker interface {
	// Start the worker
	Start(context.Context) error
	// Stop the worker
	Stop() error
	// PerformEvery a job every interval (loop)
	// if WithSync(true) Option, then the operation blocks until it's done which means only one instance can be executed at a time
	// the default is WithSync(false), so there's not blocking and you could get multiple instances running at a time if the latency is longer than the interval
	PerformEvery(*Job, time.Duration, ...Option) error
	// Perform a job as soon as possibly,  If WithSync(true) Option then it's a blocking call, the default is false (so async)
	Perform(*Job, ...Option) error
	// PerformAt performs a job at a particular time and always async
	PerformAt(*Job, time.Time) error
	// PerformIn performs a job after waiting for a specified amount of time and always async
	PerformIn(*Job, time.Duration) error
	// PeformReceive peforms a job for every value received from channel
	PerformReceive(*Job, interface{}, ...Option) error
	// Register a Handler
	Register(string, Handler) error
	// GetContext returns the worker context
	GetContext() context.Context
	// SetContext sets the worker context
	SetContext(context.Context)
}
```

## Adapter (Middleware)
Adapter defines a common func interface so middleware can be chained together for a Job.  Currently, there is adapter middleware for things like: metrics, tracing, logging, and healthchecks.  More adapters will be added as well.
``` go
type Adapter func(Handler) Handler
```
Example opentracing Adapater
```go
// WithOpenTracing is an adpater middleware that adds opentracing
func WithOpenTracing(operationPrefix []byte) Adapter {
	if operationPrefix == nil {
		operationPrefix = []byte("api-request-")
	}
	return func(h Handler) Handler {
		return Handler(func(job *Job) error {
			// all before request is handled
			var span opentracing.Span
			if cspan := job.Ctx.Value("tracing-context"); cspan != nil {
				span = StartSpanWithParent(cspan.(opentracing.Span).Context(), string(operationPrefix)+job.Handler)
			} else {
				span = StartSpan(string(operationPrefix) + job.Handler)
			}
			defer span.Finish()                  // after all the other defers are completed.. finish the span
			job.Ctx.Set("tracing-context", span) // add the span to the context so it can be used for the duration of the request.
			defer span.SetTag("status-code", job.Ctx.Status())
			err := h(job)
			return err
		})
	}
```

---
## Acknowledgements 
I need to acknowledge that many of the ideas implemented in this library are not my own.  I've been inspired and shamelessly borrowed from the following individuals/projects:
- Mat Ryer's [middleware](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81)
- Mark Bate's Buffalo [Worker](https://gobuffalo.io/en/docs/workers)

---
##  Complete API reference: [docs.md](https://github.com/Bose/go-work/blob/master/docs.md#type-commonworker)
