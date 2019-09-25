package work

import (
	"io"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerprom "github.com/uber/jaeger-lib/metrics/prometheus"
)

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
}

const optionSampleProbability = "optionSampleProbability"

// WithSampleProbability - optional sample probability
func WithSampleProbability(sampleProbability float64) Option {
	return func(o Options) {
		o[optionSampleProbability] = sampleProbability
	}
}

// InitTracing will init opentracing with options WithSampleProbability defaults: constant sampling
func InitTracing(serviceName string, tracingAgentHostPort string, opt ...Option) (
	tracer opentracing.Tracer,
	reporter jaeger.Reporter,
	closer io.Closer,
	err error) {
	opts := GetOpts(opt...)
	factory := jaegerprom.New(jaegerprom.WithRegisterer(prometheus.NewRegistry()))
	metrics := jaeger.NewMetrics(factory, map[string]string{"lib": "jaeger"})
	transport, err := jaeger.NewUDPTransport(tracingAgentHostPort, 0)
	if err != nil {
		return tracer, reporter, closer, err
	}
	reporter = jaeger.NewCompositeReporter(
		jaeger.NewRemoteReporter(transport,
			jaeger.ReporterOptions.Metrics(metrics),
		),
	)
	sampler := jaeger.NewConstSampler(true)
	if s, ok := opts[optionSampleProbability].(float64); ok && s > 0 {
		sampler, err = jaeger.NewProbabilisticSampler(s)
	}

	tracer, closer = jaeger.NewTracer(serviceName,
		sampler,
		reporter,
		jaeger.TracerOptions.Metrics(metrics),
	)
	return tracer, reporter, closer, nil
}
