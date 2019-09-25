

# work
`import "github.com/BoseCorp/go-work"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>



## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [func InitTracing(serviceName string, tracingAgentHostPort string, opt ...Option) (tracer opentracing.Tracer, reporter jaeger.Reporter, closer io.Closer, err error)](#InitTracing)
* [func RunE(fn func() error, opt ...Option) (err error)](#RunE)
* [func StartSpan(operationName string) opentracing.Span](#StartSpan)
* [func StartSpanWithParent(parent opentracing.SpanContext, operationName string) opentracing.Span](#StartSpanWithParent)
* [func WrapChannel(chanToWrap interface{}) (&lt;-chan interface{}, error)](#WrapChannel)
* [type Adapter](#Adapter)
  * [func WithAggregateLogger(useBanner bool, timeFormat string, utc bool, logrusFieldNameForTraceID string, contextTraceIDField []byte, opt ...Option) Adapter](#WithAggregateLogger)
  * [func WithHealthCheck(health *HealthCheck, opt ...Option) Adapter](#WithHealthCheck)
  * [func WithOpenTracing(operationPrefix []byte) Adapter](#WithOpenTracing)
  * [func WithPrometheus(p *Prometheus) Adapter](#WithPrometheus)
* [type Args](#Args)
* [type CommonWorker](#CommonWorker)
  * [func NewCommonWorker(l Logger) *CommonWorker](#NewCommonWorker)
  * [func NewCommonWorkerWithContext(ctx context.Context, l Logger) *CommonWorker](#NewCommonWorkerWithContext)
  * [func (w *CommonWorker) GetContext() context.Context](#CommonWorker.GetContext)
  * [func (w *CommonWorker) Perform(job *Job, opt ...Option) error](#CommonWorker.Perform)
  * [func (w *CommonWorker) PerformAt(job *Job, t time.Time) error](#CommonWorker.PerformAt)
  * [func (w *CommonWorker) PerformEvery(job *Job, interval time.Duration, opt ...Option) error](#CommonWorker.PerformEvery)
  * [func (w *CommonWorker) PerformIn(job *Job, d time.Duration) error](#CommonWorker.PerformIn)
  * [func (w *CommonWorker) PerformReceive(job *Job, readChan interface{}, opt ...Option) error](#CommonWorker.PerformReceive)
  * [func (w *CommonWorker) Register(name string, h Handler) error](#CommonWorker.Register)
  * [func (w *CommonWorker) SetContext(ctx context.Context)](#CommonWorker.SetContext)
  * [func (w *CommonWorker) Start(ctx context.Context) error](#CommonWorker.Start)
  * [func (w *CommonWorker) Stop() error](#CommonWorker.Stop)
* [type ConcurrentJob](#ConcurrentJob)
  * [func NewConcurrentJob(job Job, workerFactory WorkerFactory, performWith PerformWith, performEvery time.Duration, maxWorker int64, startInterval time.Duration, logger Logger, opt ...Option) (ConcurrentJob, error)](#NewConcurrentJob)
  * [func (j *ConcurrentJob) Register(name string, h Handler) error](#ConcurrentJob.Register)
  * [func (j *ConcurrentJob) RunningWorkers() int64](#ConcurrentJob.RunningWorkers)
  * [func (j *ConcurrentJob) Start() error](#ConcurrentJob.Start)
  * [func (j *ConcurrentJob) Stop()](#ConcurrentJob.Stop)
* [type Context](#Context)
  * [func NewContext(ctx context.Context) Context](#NewContext)
  * [func (c *Context) Get(k string) (interface{}, bool)](#Context.Get)
  * [func (c *Context) Set(k string, v interface{})](#Context.Set)
  * [func (c *Context) SetStatus(s Status)](#Context.SetStatus)
  * [func (c *Context) Status() Status](#Context.Status)
* [type Handler](#Handler)
  * [func Adapt(h Handler, adapters ...Adapter) Handler](#Adapt)
* [type HealthCheck](#HealthCheck)
  * [func NewHealthCheck(opt ...Option) *HealthCheck](#NewHealthCheck)
  * [func (h *HealthCheck) Close()](#HealthCheck.Close)
  * [func (h *HealthCheck) DefaultHealthHandler() gin.HandlerFunc](#HealthCheck.DefaultHealthHandler)
  * [func (h *HealthCheck) GetStatus() int](#HealthCheck.GetStatus)
  * [func (h *HealthCheck) SetStatus(s int)](#HealthCheck.SetStatus)
  * [func (h *HealthCheck) WithEngine(e *gin.Engine)](#HealthCheck.WithEngine)
* [type Job](#Job)
  * [func (j *Job) Copy() Job](#Job.Copy)
* [type Logger](#Logger)
* [type Metric](#Metric)
  * [func NewMetricCounter(frames ...string) Metric](#NewMetricCounter)
  * [func NewMetricStatusGauge(min int, max int, frames ...string) Metric](#NewMetricStatusGauge)
* [type Option](#Option)
  * [func WithBanner(useBanner bool) Option](#WithBanner)
  * [func WithChannel(ch interface{}) Option](#WithChannel)
  * [func WithEngine(e *gin.Engine) Option](#WithEngine)
  * [func WithErrorPercentage(percentageOfErrors float64, lastNumOfMinutes int) Option](#WithErrorPercentage)
  * [func WithErrorPercentageByRequestCount(percentageOfErrors float64, minNumOfRequests, lastNumOfRequests int) Option](#WithErrorPercentageByRequestCount)
  * [func WithHealthHandler(h gin.HandlerFunc) Option](#WithHealthHandler)
  * [func WithHealthPath(path string) Option](#WithHealthPath)
  * [func WithHealthTicker(ticker *time.Ticker) Option](#WithHealthTicker)
  * [func WithIgnore(handlers ...string) Option](#WithIgnore)
  * [func WithJob(j *Job) Option](#WithJob)
  * [func WithMetricsPath(path string) Option](#WithMetricsPath)
  * [func WithNamespace(ns string) Option](#WithNamespace)
  * [func WithSampleProbability(sampleProbability float64) Option](#WithSampleProbability)
  * [func WithSilentNoResponse(silent bool) Option](#WithSilentNoResponse)
  * [func WithSubSystem(sub string) Option](#WithSubSystem)
  * [func WithSync(sync bool) Option](#WithSync)
* [type Options](#Options)
  * [func GetOpts(opt ...Option) Options](#GetOpts)
  * [func (o *Options) Get(name string) (interface{}, bool)](#Options.Get)
* [type PerformWith](#PerformWith)
  * [func (p PerformWith) String() string](#PerformWith.String)
* [type Prometheus](#Prometheus)
  * [func NewPrometheus(opt ...Option) *Prometheus](#NewPrometheus)
  * [func (p *Prometheus) WithEngine(e *gin.Engine)](#Prometheus.WithEngine)
* [type Session](#Session)
  * [func NewSession() Session](#NewSession)
  * [func (c *Session) Get(k string) (interface{}, bool)](#Session.Get)
  * [func (c *Session) Set(k string, v interface{})](#Session.Set)
* [type Status](#Status)
* [type Worker](#Worker)
  * [func NewCommonWorkerFactory(ctx context.Context, l Logger) Worker](#NewCommonWorkerFactory)
* [type WorkerFactory](#WorkerFactory)


#### <a name="pkg-files">Package files</a>
[channel.go](/src/github.com/BoseCorp/go-work/channel.go) [concurrent_job.go](/src/github.com/BoseCorp/go-work/concurrent_job.go) [context.go](/src/github.com/BoseCorp/go-work/context.go) [job.go](/src/github.com/BoseCorp/go-work/job.go) [logger.go](/src/github.com/BoseCorp/go-work/logger.go) [metrics.go](/src/github.com/BoseCorp/go-work/metrics.go) [middleware.go](/src/github.com/BoseCorp/go-work/middleware.go) [middleware_aggregatelogger.go](/src/github.com/BoseCorp/go-work/middleware_aggregatelogger.go) [middleware_health.go](/src/github.com/BoseCorp/go-work/middleware_health.go) [middleware_opentracing.go](/src/github.com/BoseCorp/go-work/middleware_opentracing.go) [middleware_prometheus.go](/src/github.com/BoseCorp/go-work/middleware_prometheus.go) [options.go](/src/github.com/BoseCorp/go-work/options.go) [safe.go](/src/github.com/BoseCorp/go-work/safe.go) [session.go](/src/github.com/BoseCorp/go-work/session.go) [span.go](/src/github.com/BoseCorp/go-work/span.go) [status.go](/src/github.com/BoseCorp/go-work/status.go) [worker.go](/src/github.com/BoseCorp/go-work/worker.go) [worker_common.go](/src/github.com/BoseCorp/go-work/worker_common.go) 


## <a name="pkg-constants">Constants</a>
``` go
const (

    // RequestSize is the index used by Job.Ctx.Set(string, interface{}) or Job.Ctx.Get(string) to communicate the request approximate size in bytes
    // WithPrometheus optionally will report this info when RequestSize is found in the Job.Ctxt
    RequestSize = "keyRequestSize"

    // ResponseSize is the index used by Job.Ctx.Set(string, interface{}) or Job.Ctx.Get(string) to communicate the response approximate size in bytes
    // WithPrometheus optionally will report this info when RequestSize is found in the Job.Ctxt
    ResponseSize = "keyResponseSize"
)
```
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
``` go
const AggregateLogger = "aggregateLogger"
```
AggregateLogger defines the const string for getting the logger from a Job context

``` go
const DefaultHealthTickerDuration = 1 * time.Minute
```
DefaultHealthTickerDuration is the time duration between the recalculation of the status returned by HealthCheck.GetStatus()


## <a name="pkg-variables">Variables</a>
``` go
var ContextTraceIDField string
```
ContextTraceIDField - used to find the trace id in the context - optional



## <a name="InitTracing">func</a> [InitTracing](./middleware_opentracing.go?s=1497:1665#L45)
``` go
func InitTracing(serviceName string, tracingAgentHostPort string, opt ...Option) (
    tracer opentracing.Tracer,
    reporter jaeger.Reporter,
    closer io.Closer,
    err error)
```
InitTracing will init opentracing with options WithSampleProbability defaults: constant sampling



## <a name="RunE">func</a> [RunE](./safe.go?s=434:487#L24)
``` go
func RunE(fn func() error, opt ...Option) (err error)
```
Run the function safely knowing that if it panics,
the panic will be caught and returned as an error.
it produces a stack trace and if WithJob(job), the job's status is set to StatusInternalError.



## <a name="StartSpan">func</a> [StartSpan](./span.go?s=192:245#L11)
``` go
func StartSpan(operationName string) opentracing.Span
```
StartSpan will start a new span with no parent span.



## <a name="StartSpanWithParent">func</a> [StartSpanWithParent](./span.go?s=437:532#L18)
``` go
func StartSpanWithParent(parent opentracing.SpanContext, operationName string) opentracing.Span
```
StartSpanWithParent will start a new span with a parent span.
example:


	span:= StartSpanWithParent(c.Get("tracing-context"),



## <a name="WrapChannel">func</a> [WrapChannel](./channel.go?s=206:274#L10)
``` go
func WrapChannel(chanToWrap interface{}) (<-chan interface{}, error)
```
WrapChanel takes a concrete receiving chan in as an interface{}, and wraps it with an interface{} chan
so you can treat all receiving channels the same way




## <a name="Adapter">type</a> [Adapter](./middleware.go?s=309:343#L11)
``` go
type Adapter func(Handler) Handler
```
Adapter defines the adaptor middleware type







### <a name="WithAggregateLogger">func</a> [WithAggregateLogger](./middleware_aggregatelogger.go?s=1020:1181#L37)
``` go
func WithAggregateLogger(
    useBanner bool,
    timeFormat string,
    utc bool,
    logrusFieldNameForTraceID string,
    contextTraceIDField []byte,
    opt ...Option) Adapter
```
WithAggregateLogger is a middleware adapter for aggregated logging (see go-gin-logrus)


### <a name="WithHealthCheck">func</a> [WithHealthCheck](./middleware_health.go?s=407:471#L17)
``` go
func WithHealthCheck(health *HealthCheck, opt ...Option) Adapter
```
WithHealthCheck is an adpater middleware for healthcheck.
it also adds a health http GET endpoint.  It supports the Option.
WithErrorPercentage(percentageOfErrors float64, lastNumOfMinutes int) that allows
you to override the default of: 1.0 (100%) errors in the last 5 min.


### <a name="WithOpenTracing">func</a> [WithOpenTracing](./middleware_opentracing.go?s=305:357#L13)
``` go
func WithOpenTracing(operationPrefix []byte) Adapter
```
WithOpenTracing is an adpater middleware that adds opentracing


### <a name="WithPrometheus">func</a> [WithPrometheus](./middleware_prometheus.go?s=962:1004#L32)
``` go
func WithPrometheus(p *Prometheus) Adapter
```
Instrument is a gin middleware that can be used to generate metrics for a
single handler





## <a name="Args">type</a> [Args](./middleware.go?s=59:91#L4)
``` go
type Args map[string]interface{}
```
Args is how parameters are passed to jobs










## <a name="CommonWorker">type</a> [CommonWorker](./worker_common.go?s=135:622#L13)
``` go
type CommonWorker struct {

    // Logger for the worker
    Logger Logger
    // contains filtered or unexported fields
}

```
CommonWorker defines the typical common worker







### <a name="NewCommonWorker">func</a> [NewCommonWorker](./worker_common.go?s=702:746#L37)
``` go
func NewCommonWorker(l Logger) *CommonWorker
```
NewCommonWorker creates a new CommonWorker


### <a name="NewCommonWorkerWithContext">func</a> [NewCommonWorkerWithContext](./worker_common.go?s=869:945#L42)
``` go
func NewCommonWorkerWithContext(ctx context.Context, l Logger) *CommonWorker
```
NewCommonWorkerWithContext creates a new CommonWorker





### <a name="CommonWorker.GetContext">func</a> (\*CommonWorker) [GetContext](./worker_common.go?s=1834:1885#L76)
``` go
func (w *CommonWorker) GetContext() context.Context
```
GetContext from the worker




### <a name="CommonWorker.Perform">func</a> (\*CommonWorker) [Perform](./worker_common.go?s=2121:2182#L86)
``` go
func (w *CommonWorker) Perform(job *Job, opt ...Option) error
```
Perform executes the job.  If WithSync(true) Option then it's a blocking call, the default is false (so async)




### <a name="CommonWorker.PerformAt">func</a> (\*CommonWorker) [PerformAt](./worker_common.go?s=6957:7018#L236)
``` go
func (w *CommonWorker) PerformAt(job *Job, t time.Time) error
```
PerformAt performs a job at a particular time using a goroutine.




### <a name="CommonWorker.PerformEvery">func</a> (\*CommonWorker) [PerformEvery](./worker_common.go?s=4058:4148#L148)
``` go
func (w *CommonWorker) PerformEvery(job *Job, interval time.Duration, opt ...Option) error
```
PerformEvery executes the job on the interval.
if WithSync(true) Option, then the operation blocks until it's done which means only one instance can be executed at a time.
the default is WithSync(false), so there's not blocking and you could get multiple instances running at a time if the latency is longer than the interval.




### <a name="CommonWorker.PerformIn">func</a> (\*CommonWorker) [PerformIn](./worker_common.go?s=6374:6439#L215)
``` go
func (w *CommonWorker) PerformIn(job *Job, d time.Duration) error
```
PerformIn performs a job after the "in" time has expired.




### <a name="CommonWorker.PerformReceive">func</a> (\*CommonWorker) [PerformReceive](./worker_common.go?s=5266:5356#L179)
``` go
func (w *CommonWorker) PerformReceive(job *Job, readChan interface{}, opt ...Option) error
```
PerformReceive will loop on receiving data from the readChan (passed as a simple interface{}).
This uses the Job ctx for timeouts and cancellation (just like the rest of the framework)




### <a name="CommonWorker.Register">func</a> (\*CommonWorker) [Register](./worker_common.go?s=7198:7259#L244)
``` go
func (w *CommonWorker) Register(name string, h Handler) error
```
Register Handler with the worker




### <a name="CommonWorker.SetContext">func</a> (\*CommonWorker) [SetContext](./worker_common.go?s=1934:1988#L81)
``` go
func (w *CommonWorker) SetContext(ctx context.Context)
```
SetContext for the worker




### <a name="CommonWorker.Start">func</a> (\*CommonWorker) [Start](./worker_common.go?s=1678:1733#L70)
``` go
func (w *CommonWorker) Start(ctx context.Context) error
```
Start the worker




### <a name="CommonWorker.Stop">func</a> (\*CommonWorker) [Stop](./worker_common.go?s=1481:1516#L61)
``` go
func (w *CommonWorker) Stop() error
```
Stop the worker and you must call Stop() to clean up the CommonWorker internal Ctx (or it will leak memory)




## <a name="ConcurrentJob">type</a> [ConcurrentJob](./concurrent_job.go?s=762:1879#L45)
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

    PerformReceiveChan interface{}
    // contains filtered or unexported fields
}

```
ConcurrentJob represents a job to be run concurrently







### <a name="NewConcurrentJob">func</a> [NewConcurrentJob](./concurrent_job.go?s=1917:2139#L90)
``` go
func NewConcurrentJob(
    job Job,
    workerFactory WorkerFactory,
    performWith PerformWith,
    performEvery time.Duration,
    maxWorker int64,
    startInterval time.Duration,
    logger Logger,
    opt ...Option,
) (ConcurrentJob, error)
```
NewConcurrentJob makes a new job





### <a name="ConcurrentJob.Register">func</a> (\*ConcurrentJob) [Register](./concurrent_job.go?s=3213:3275#L129)
``` go
func (j *ConcurrentJob) Register(name string, h Handler) error
```
Register a handler for the Job's workers




### <a name="ConcurrentJob.RunningWorkers">func</a> (\*ConcurrentJob) [RunningWorkers](./concurrent_job.go?s=3656:3702#L144)
``` go
func (j *ConcurrentJob) RunningWorkers() int64
```



### <a name="ConcurrentJob.Start">func</a> (\*ConcurrentJob) [Start](./concurrent_job.go?s=6001:6038#L213)
``` go
func (j *ConcurrentJob) Start() error
```
Start all the work




### <a name="ConcurrentJob.Stop">func</a> (\*ConcurrentJob) [Stop](./concurrent_job.go?s=5803:5833#L204)
``` go
func (j *ConcurrentJob) Stop()
```
Stop all the work and you must call Stop() to clean up the ConcurrentJob Ctx (or it will leak memory)




## <a name="Context">type</a> [Context](./context.go?s=101:424#L9)
``` go
type Context struct {
    // context.Context allows it to be a compatible context
    context.Context
    // contains filtered or unexported fields
}

```
Context for the Job and is reset for every execution







### <a name="NewContext">func</a> [NewContext](./context.go?s=448:492#L24)
``` go
func NewContext(ctx context.Context) Context
```
NewContext factory





### <a name="Context.Get">func</a> (\*Context) [Get](./context.go?s=595:646#L34)
``` go
func (c *Context) Get(k string) (interface{}, bool)
```
Get a key/value




### <a name="Context.Set">func</a> (\*Context) [Set](./context.go?s=759:805#L42)
``` go
func (c *Context) Set(k string, v interface{})
```
Set a key/value




### <a name="Context.SetStatus">func</a> (\*Context) [SetStatus](./context.go?s=990:1027#L54)
``` go
func (c *Context) SetStatus(s Status)
```
SetStatus for the context




### <a name="Context.Status">func</a> (\*Context) [Status](./context.go?s=905:938#L49)
``` go
func (c *Context) Status() Status
```
Status retrieves the context's status




## <a name="Handler">type</a> [Handler](./middleware.go?s=229:260#L8)
``` go
type Handler func(j *Job) error
```
Handler is executed a a Work for a given Job.  It also defines the interface
for handlers that can be used by middleware adapters







### <a name="Adapt">func</a> [Adapt](./middleware.go?s=397:447#L14)
``` go
func Adapt(h Handler, adapters ...Adapter) Handler
```
Adapt a handler with provided middlware adapters





## <a name="HealthCheck">type</a> [HealthCheck](./middleware_health.go?s=3678:4116#L134)
``` go
type HealthCheck struct {

    // HealthPath is the GET url path for the endpoint
    HealthPath string

    // Engine is the gin.Engine that should serve the endpoint
    Engine *gin.Engine

    // Handler is the gin Hanlder to use for the endpoint
    Handler gin.HandlerFunc
    // contains filtered or unexported fields
}

```
HealthCheck provides a healthcheck endpoint for the work







### <a name="NewHealthCheck">func</a> [NewHealthCheck](./middleware_health.go?s=4315:4362#L157)
``` go
func NewHealthCheck(opt ...Option) *HealthCheck
```
NewHealthCheck creates a new HealthCheck with the options provided.
Options: WithEngine(*gin.Engine), WithHealthPath(string), WithHealthHander(gin.HandlerFunc), WithMetricTicker(time.Ticker)





### <a name="HealthCheck.Close">func</a> (\*HealthCheck) [Close](./middleware_health.go?s=5089:5118#L190)
``` go
func (h *HealthCheck) Close()
```
Close cleans up the all the HealthCheck resources




### <a name="HealthCheck.DefaultHealthHandler">func</a> (\*HealthCheck) [DefaultHealthHandler](./middleware_health.go?s=5541:5601#L209)
``` go
func (h *HealthCheck) DefaultHealthHandler() gin.HandlerFunc
```



### <a name="HealthCheck.GetStatus">func</a> (\*HealthCheck) [GetStatus](./middleware_health.go?s=5194:5231#L195)
``` go
func (h *HealthCheck) GetStatus() int
```
GetStatus returns the current health status




### <a name="HealthCheck.SetStatus">func</a> (\*HealthCheck) [SetStatus](./middleware_health.go?s=5310:5348#L200)
``` go
func (h *HealthCheck) SetStatus(s int)
```
SetStatus sets the current health status




### <a name="HealthCheck.WithEngine">func</a> (\*HealthCheck) [WithEngine](./middleware_health.go?s=5474:5521#L205)
``` go
func (h *HealthCheck) WithEngine(e *gin.Engine)
```
WithEngine lets you set the *gin.Engine if it's created after you've created the *HealthCheck




## <a name="Job">type</a> [Job](./job.go?s=69:448#L8)
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
Job to be processed by a Worker










### <a name="Job.Copy">func</a> (\*Job) [Copy](./job.go?s=490:514#L26)
``` go
func (j *Job) Copy() Job
```
Copy provides a deep copy of the Job




## <a name="Logger">type</a> [Logger](./logger.go?s=56:244#L4)
``` go
type Logger interface {
    Debugf(string, ...interface{})
    Infof(string, ...interface{})
    Errorf(string, ...interface{})
    Debug(...interface{})
    Info(...interface{})
    Error(...interface{})
}
```
Logger is used by worker to write logs










## <a name="Metric">type</a> [Metric](./metrics.go?s=267:342#L18)
``` go
type Metric interface {
    Add(n float64)
    String() string
    Value() float64
}
```
Metric is a single meter (a counter for now, but in the future: gauge or histogram, optionally - with history)







### <a name="NewMetricCounter">func</a> [NewMetricCounter](./metrics.go?s=725:771#L36)
``` go
func NewMetricCounter(frames ...string) Metric
```
NewCounter returns a counter metric that increments the value with each
incoming number.


### <a name="NewMetricStatusGauge">func</a> [NewMetricStatusGauge](./metrics.go?s=4212:4280#L206)
``` go
func NewMetricStatusGauge(min int, max int, frames ...string) Metric
```
NewMetricStatusGauge is a factory for statusGauge Metrics





## <a name="Option">type</a> [Option](./options.go?s=244:269#L13)
``` go
type Option func(Options)
```
Option - how Options are passed as arguments







### <a name="WithBanner">func</a> [WithBanner](./middleware_aggregatelogger.go?s=540:578#L21)
``` go
func WithBanner(useBanner bool) Option
```
WithBanner specifys the table name to use for an outbox


### <a name="WithChannel">func</a> [WithChannel](./options.go?s=856:895#L45)
``` go
func WithChannel(ch interface{}) Option
```
WithChannel optional channel parameter


### <a name="WithEngine">func</a> [WithEngine](./middleware_prometheus.go?s=3587:3624#L134)
``` go
func WithEngine(e *gin.Engine) Option
```
WithEngine is an option allowing to set the gin engine when intializing with New.
Example :
r := gin.Default()
p := work.NewPrometheus(WithEngine(r))


### <a name="WithErrorPercentage">func</a> [WithErrorPercentage](./middleware_health.go?s=7022:7103#L260)
``` go
func WithErrorPercentage(percentageOfErrors float64, lastNumOfMinutes int) Option
```
WithErrorPercentage allows you to override the default of 1.0 (100%) with the % you want for error rate.


### <a name="WithErrorPercentageByRequestCount">func</a> [WithErrorPercentageByRequestCount](./middleware_health.go?s=7352:7466#L271)
``` go
func WithErrorPercentageByRequestCount(percentageOfErrors float64, minNumOfRequests, lastNumOfRequests int) Option
```

### <a name="WithHealthHandler">func</a> [WithHealthHandler](./middleware_health.go?s=6453:6501#L241)
``` go
func WithHealthHandler(h gin.HandlerFunc) Option
```
WithHealthHandler override the default health endpoint handler


### <a name="WithHealthPath">func</a> [WithHealthPath](./middleware_health.go?s=6690:6729#L250)
``` go
func WithHealthPath(path string) Option
```
WithHealthPath override the default path for the health endpoint


### <a name="WithHealthTicker">func</a> [WithHealthTicker](./middleware_health.go?s=6211:6260#L232)
``` go
func WithHealthTicker(ticker *time.Ticker) Option
```

### <a name="WithIgnore">func</a> [WithIgnore](./middleware_prometheus.go?s=4124:4166#L153)
``` go
func WithIgnore(handlers ...string) Option
```
WithIgnore is used to disable instrumentation on some routes


### <a name="WithJob">func</a> [WithJob](./options.go?s=682:709#L36)
``` go
func WithJob(j *Job) Option
```
WithJob optional Job parameter


### <a name="WithMetricsPath">func</a> [WithMetricsPath](./middleware_prometheus.go?s=3890:3930#L144)
``` go
func WithMetricsPath(path string) Option
```
WithMetricsPath is an option allowing to set the metrics path when intializing with New.
Example : work.New(work.WithMetricsPath("/mymetrics"))


### <a name="WithNamespace">func</a> [WithNamespace](./middleware_prometheus.go?s=4744:4780#L175)
``` go
func WithNamespace(ns string) Option
```
WithNamespace is an option allowing to set the namespace when intitializing
with New.
Example : work.New(work.WithNamespace("my_namespace"))


### <a name="WithSampleProbability">func</a> [WithSampleProbability](./middleware_opentracing.go?s=1253:1313#L38)
``` go
func WithSampleProbability(sampleProbability float64) Option
```
WithSampleProbability - optional sample probability


### <a name="WithSilentNoResponse">func</a> [WithSilentNoResponse](./middleware_aggregatelogger.go?s=813:858#L30)
``` go
func WithSilentNoResponse(silent bool) Option
```
WithSilentNoReponse specifies that StatusNoResponse requests should be silent (no logging)


### <a name="WithSubSystem">func</a> [WithSubSystem](./middleware_prometheus.go?s=4440:4477#L164)
``` go
func WithSubSystem(sub string) Option
```
WithSubsystem is an option allowing to set the subsystem when intitializing
with New.
Example : work.New(work.WithSubsystem("my_system"))


### <a name="WithSync">func</a> [WithSync](./options.go?s=516:547#L27)
``` go
func WithSync(sync bool) Option
```
WithSync optional synchronous execution





## <a name="Options">type</a> [Options](./options.go?s=312:347#L16)
``` go
type Options map[string]interface{}
```
Options = how options are represented







### <a name="GetOpts">func</a> [GetOpts](./options.go?s=75:110#L4)
``` go
func GetOpts(opt ...Option) Options
```
GetOpts - iterate the inbound Options and return a struct





### <a name="Options.Get">func</a> (\*Options) [Get](./options.go?s=991:1045#L52)
``` go
func (o *Options) Get(name string) (interface{}, bool)
```
Get a specific option by name




## <a name="PerformWith">type</a> [PerformWith](./concurrent_job.go?s=130:150#L16)
``` go
type PerformWith int
```

``` go
const (
    PerformWithUnknown PerformWith = iota
    PerformWithSync
    PerformWithAsync
    PerformEveryWithSync
    PerformEveryWithAsync
    PerformReceiveWithSync
    PerformReceiveWithAsync
)
```









### <a name="PerformWith.String">func</a> (PerformWith) [String](./concurrent_job.go?s=331:367#L28)
``` go
func (p PerformWith) String() string
```



## <a name="Prometheus">type</a> [Prometheus](./middleware_prometheus.go?s=2023:2247#L72)
``` go
type Prometheus struct {
    MetricsPath string
    Namespace   string
    Subsystem   string
    Ignored     isPresentMap
    Engine      *gin.Engine
    // contains filtered or unexported fields
}

```
Prometheus contains the metrics gathered by the instance and its path







### <a name="NewPrometheus">func</a> [NewPrometheus](./middleware_prometheus.go?s=2475:2520#L87)
``` go
func NewPrometheus(opt ...Option) *Prometheus
```
New will initialize a new Prometheus instance with the given options.
If no options are passed, sane defaults are used.
If a router is passed using the Engine() option, this instance will
automatically bind to it.





### <a name="Prometheus.WithEngine">func</a> (\*Prometheus) [WithEngine](./middleware_prometheus.go?s=4947:4993#L183)
``` go
func (p *Prometheus) WithEngine(e *gin.Engine)
```
WithEngine is a method that should be used if the engine is set after middleware
initialization




## <a name="Session">type</a> [Session](./session.go?s=182:295#L9)
``` go
type Session struct {
    // contains filtered or unexported fields
}

```
Session that is used to pass session info to a Job
this is a good spot to put things like *redis.Pool or *sqlx.DB for outbox connection pools







### <a name="NewSession">func</a> [NewSession](./session.go?s=319:344#L16)
``` go
func NewSession() Session
```
NewSession factory





### <a name="Session.Get">func</a> (\*Session) [Get](./session.go?s=434:485#L24)
``` go
func (c *Session) Get(k string) (interface{}, bool)
```
Get a key/value




### <a name="Session.Set">func</a> (\*Session) [Set](./session.go?s=598:644#L32)
``` go
func (c *Session) Set(k string, v interface{})
```
Set a key/value




## <a name="Status">type</a> [Status](./status.go?s=56:71#L4)
``` go
type Status int
```
Status for a job's execution (Perform)










## <a name="Worker">type</a> [Worker](./worker.go?s=45:1242#L8)
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






### <a name="NewCommonWorkerFactory">func</a> [NewCommonWorkerFactory](./worker_common.go?s=1238:1303#L55)
``` go
func NewCommonWorkerFactory(ctx context.Context, l Logger) Worker
```
NewCommonWorkerFactory is a simple adapter that allows the factory to conform to the WorkerFactory type





## <a name="WorkerFactory">type</a> [WorkerFactory](./concurrent_job.go?s=648:703#L42)
``` go
type WorkerFactory func(context.Context, Logger) Worker
```













- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
