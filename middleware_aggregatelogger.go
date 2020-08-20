package work

import (
	"fmt"
	"time"

	ginlogrus "github.com/Bose/go-gin-logrus/v2"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

// ContextTraceIDField - used to find the trace id in the context - optional
var ContextTraceIDField string

// AggregateLogger defines the const string for getting the logger from a Job context
const AggregateLogger = "aggregateLogger"

const optionAggregateLoggingUseBanner = "optionAggregateLoggingUseBanner"

// WithBanner specifies the table name to use for an outbox
func WithBanner(useBanner bool) Option {
	return func(o Options) {
		o[optionAggregateLoggingUseBanner] = useBanner
	}
}

const optionSilentNoResponse = "optionSilentNoResponse"

// WithSilentNoResponse specifies that StatusNoResponse requests should be silent (no logging)
func WithSilentNoResponse(silent bool) Option {
	return func(o Options) {
		o[optionSilentNoResponse] = silent
	}
}

const optionSilentSuccess = "optionSilentSuccess"
// WithSilentSuccess specifies that StatusSuccess requests should be silent (no logging)
func WithSilentSuccess(silent bool) Option {
	return func(o Options) {
		o[optionSilentSuccess] = silent
	}
}

// ReducedLoggingFunc defines a function type used for custom logic on when to print logs
type ReducedLoggingFunc func(workStatus Status, logBufferLength int) bool

var DefaultReducedLoggingFunc ReducedLoggingFunc = func(s Status, l int) bool {return false}

const optionReducedLoggingFunc = "optionReducedLoggingFunc"
// WithReducedLoggingFunc specifies the function used to set custom logic around when to print logs
func WithReducedLoggingFunc(a ReducedLoggingFunc) Option {
	return func(o Options) {
		o[optionReducedLoggingFunc] = a
	}
}

const optionLogLevel = "optionLogLevel"
// WithLogLevel will set the logrus log level for the job handler
func WithLogLevel(level logrus.Level) Option {
	return func(o Options) {
		o[optionLogLevel] = level
	}
}

// WithAggregateLogger is a middleware adapter for aggregated logging (see go-gin-logrus)
func WithAggregateLogger(
	useBanner bool,
	timeFormat string,
	utc bool,
	logrusFieldNameForTraceID string,
	contextTraceIDField []byte,
	opt ...Option) Adapter {
	opts := GetOpts(opt...)

	if contextTraceIDField != nil {
		ContextTraceIDField = string(contextTraceIDField)
	}

	return func(h Handler) Handler {
		return Handler(func(job *Job) error {
			useBanner := false
			if b, ok := opts[optionAggregateLoggingUseBanner].(bool); ok {
				useBanner = b
			}
			silentNoResponse := false
			if b, ok := opts[optionSilentNoResponse].(bool); ok {
				silentNoResponse = b
			}
			silentSuccess := false
			if b, ok := opts[optionSilentSuccess].(bool); ok {
				silentSuccess = b
			}
			var reducedLoggingFunc ReducedLoggingFunc = DefaultReducedLoggingFunc
			if f, ok := opts[optionReducedLoggingFunc].(ReducedLoggingFunc); ok {
				reducedLoggingFunc = f
			}
			logLevel := logrus.DebugLevel
			if l, ok := opts[optionLogLevel].(logrus.Level); ok {
				logLevel = l
			}
			aggregateLoggingBuff := ginlogrus.NewLogBuffer(ginlogrus.WithBanner(false))
			aggregateRequestLogger := &logrus.Logger{
				Out:       &aggregateLoggingBuff,
				Formatter: new(logrus.JSONFormatter),
				Hooks:     make(logrus.LevelHooks),
				Level:     logLevel,
			}

			start := time.Now()

			// you have to use this logger for every *logrus.Entry you create
			job.Ctx.Set(AggregateLogger, aggregateRequestLogger)

			// this will be deferred until after the chained handlers are executed
			defer func() {
				if (silentNoResponse && job.Ctx.Status() == StatusNoResponse) ||
					(silentSuccess && job.Ctx.Status() == StatusSuccess) ||
					reducedLoggingFunc(job.Ctx.Status(), aggregateLoggingBuff.Length()) {
					return
				}

				end := time.Now()
				latency := end.Sub(start)
				if utc {
					end = end.UTC()
				}

				var requestID string
				// see if we're using github.com/Bose/go-gin-opentracing which will set a span in "tracing-context"
				if s, ok := job.Ctx.Get("tracing-context"); ok {
					span := s.(opentracing.Span)
					requestID = fmt.Sprintf("%v", span)
				}
				// check a user defined context field
				if len(requestID) == 0 && contextTraceIDField != nil {
					if id, found := job.Ctx.Get(string(ContextTraceIDField)); found {
						requestID = id.(string)
					}
				}

				var comment interface{}
				if c, found := job.Ctx.Get("comment"); found {
					comment = c
				}
				fields := logrus.Fields{
					logrusFieldNameForTraceID: requestID,
					"status":                  job.Ctx.Status(),
					"handler":                 job.Handler,
					"latency-ms":              float64(latency) / float64(time.Millisecond),
					"time":                    end.Format(timeFormat),
				}
				if workerNumber, ok := job.Ctx.Value("workerNumber").(int64); ok {
					fields["workerNumber"] = workerNumber
				}
				if comment != nil {
					fields["comment"] = comment
				}
				aggregateLoggingBuff.StoreHeader("request-summary-info", fields)
				if useBanner {
					// need to upgrade go-gin-logrus to let clients define their own banner
					// aggregateLoggingBuff.CustomBanner := "-------------------------------------------------------------"
					aggregateLoggingBuff.AddBanner = true
				}
				fmt.Printf(aggregateLoggingBuff.String())

			}()

			err := h(job)
			return err
		})
	}
}
