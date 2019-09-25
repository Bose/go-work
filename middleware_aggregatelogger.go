package work

import (
	"fmt"
	"time"

	ginlogrus "github.com/Bose/go-gin-logrus"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

// ContextTraceIDField - used to find the trace id in the context - optional
var ContextTraceIDField string

// AggregateLogger defines the const string for getting the logger from a Job context
const AggregateLogger = "aggregateLogger"

const optionAggregateLoggingUseBanner = "optionAggregateLoggingUseBanner"

// WithBanner specifys the table name to use for an outbox
func WithBanner(useBanner bool) Option {
	return func(o Options) {
		o[optionAggregateLoggingUseBanner] = useBanner
	}
}

const optionSilentNoResponse = "optionSilentNoResponse"

// WithSilentNoReponse specifies that StatusNoResponse requests should be silent (no logging)
func WithSilentNoResponse(silent bool) Option {
	return func(o Options) {
		o[optionSilentNoResponse] = silent
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
			aggregateLoggingBuff := ginlogrus.NewLogBuffer(ginlogrus.WithBanner(false))
			aggregateRequestLogger := &logrus.Logger{
				Out:       &aggregateLoggingBuff,
				Formatter: new(logrus.JSONFormatter),
				Hooks:     make(logrus.LevelHooks),
				Level:     logrus.DebugLevel,
			}

			start := time.Now()

			// you have to use this logger for every *logrus.Entry you create
			job.Ctx.Set(AggregateLogger, aggregateRequestLogger)

			// this will be deferred until after the chained handlers are exectuted
			defer func() {
				if silentNoResponse && job.Ctx.Status() == StatusNoResponse {
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
					// need to upgrade go-gin-logurs to let clients define their own banner
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
