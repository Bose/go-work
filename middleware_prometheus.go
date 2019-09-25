package work

import (
	//	"strconv"

	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (

	// RequestSize is the index used by Job.Ctx.Set(string, interface{}) or Job.Ctx.Get(string) to communicate the request approximate size in bytes
	// WithPrometheus optionally will report this info when RequestSize is found in the Job.Ctxt
	RequestSize = "keyRequestSize"

	// ResponseSize is the index used by Job.Ctx.Set(string, interface{}) or Job.Ctx.Get(string) to communicate the response approximate size in bytes
	// WithPrometheus optionally will report this info when RequestSize is found in the Job.Ctxt
	ResponseSize = "keyResponseSize"

	defaultPath      = "/metrics"
	defaultNamespace = "goWork"
	defaultSubSystem = "defaultSubSystem"
)

// Instrument is a gin middleware that can be used to generate metrics for a
// single handler
func WithPrometheus(p *Prometheus) Adapter {
	return func(h Handler) Handler {
		return Handler(func(job *Job) error {
			if p.shouldIgnore(job.Handler) {
				err := h(job)
				return err
			}
			start := time.Now()
			defer func() {
				elapsed := float64(time.Since(start)) / float64(time.Second)
				status := strconv.Itoa(int(job.Ctx.Status()))

				p.reqDur.Observe(elapsed)
				p.reqCnt.WithLabelValues(status, job.Handler).Inc()

				// optionally report request size (if set by Job)
				if requestSize, found := job.Ctx.Get(RequestSize); found {
					if bytes, ok := requestSize.(float64); ok {
						p.reqSz.Observe(float64(bytes))
					}
				}
				// optionally report response size (if set by Job)
				if respSize, found := job.Ctx.Get(ResponseSize); found {
					if bytes, ok := respSize.(float64); ok {
						p.resSz.Observe(float64(bytes))
					}
				}
			}()
			err := h(job)
			return err
		})
	}
}

type isPresentMap struct {
	sync.RWMutex
	values map[string]struct{}
}

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqCnt               *prometheus.CounterVec
	reqDur, reqSz, resSz prometheus.Summary

	MetricsPath string
	Namespace   string
	Subsystem   string
	Ignored     isPresentMap
	Engine      *gin.Engine
}

// New will initialize a new Prometheus instance with the given options.
// If no options are passed, sane defaults are used.
// If a router is passed using the Engine() option, this instance will
// automatically bind to it.
func NewPrometheus(opt ...Option) *Prometheus {
	opts := GetOpts(opt...)

	p := &Prometheus{
		MetricsPath: defaultPath,
		Namespace:   defaultNamespace,
		Subsystem:   defaultSubSystem,
	}
	p.Ignored.values = make(map[string]struct{})

	if e, ok := opts[optionWithEngine].(*gin.Engine); ok {
		p.Engine = e
	}
	if path, ok := opts[optionWithMetricsPath].(string); ok {
		p.MetricsPath = path
	} else {
		p.MetricsPath = "/metrics"
	}
	if handlers, ok := opts[optionWithIgnoreHandlers].([]string); ok {
		p.Ignored.Lock()
		defer p.Ignored.Unlock()
		for _, handlerName := range handlers {
			p.Ignored.values[handlerName] = struct{}{}
		}
	}

	if sub, ok := opts[optionWithSubSystem].(string); ok {
		p.Subsystem = sub
	}
	if ns, ok := opts[optionWithNamespace].(string); ok {
		p.Namespace = ns
	}

	p.register()
	if p.Engine != nil {
		p.Engine.GET(p.MetricsPath, prometheusHandler())
	}

	return p
}

const optionWithEngine = "optionWithEngine"

// WithEngine is an option allowing to set the gin engine when intializing with New.
// Example :
// r := gin.Default()
// p := work.NewPrometheus(WithEngine(r))
func WithEngine(e *gin.Engine) Option {
	return func(o Options) {
		o[optionWithEngine] = e
	}
}

const optionWithMetricsPath = "optionWithMetricsPath"

// WithMetricsPath is an option allowing to set the metrics path when intializing with New.
// Example : work.New(work.WithMetricsPath("/mymetrics"))
func WithMetricsPath(path string) Option {
	return func(o Options) {
		o[optionWithMetricsPath] = path
	}
}

const optionWithIgnoreHandlers = "optionWithIgnoreHandlers"

// WithIgnore is used to disable instrumentation on some routes
func WithIgnore(handlers ...string) Option {
	return func(o Options) {
		o[optionWithIgnoreHandlers] = handlers
	}
}

const optionWithSubSystem = "optionWithSubSystem"

// WithSubsystem is an option allowing to set the subsystem when intitializing
// with New.
// Example : work.New(work.WithSubsystem("my_system"))
func WithSubSystem(sub string) Option {
	return func(o Options) {
		o[optionWithSubSystem] = sub
	}
}

const optionWithNamespace = "optionWithNamespace"

// WithNamespace is an option allowing to set the namespace when intitializing
// with New.
// Example : work.New(work.WithNamespace("my_namespace"))
func WithNamespace(ns string) Option {
	return func(o Options) {
		o[optionWithNamespace] = ns
	}
}

// WithEngine is a method that should be used if the engine is set after middleware
// initialization
func (p *Prometheus) WithEngine(e *gin.Engine) {
	e.GET(p.MetricsPath, prometheusHandler())
	p.Engine = e
}

func (p *Prometheus) register() {
	p.reqCnt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: p.Namespace,
			Subsystem: p.Subsystem,
			Name:      "requests_total",
			Help:      "How work requests processed, partitioned by job status code",
		},
		[]string{"code", "handler"},
	)
	prometheus.MustRegister(p.reqCnt)

	p.reqDur = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: p.Namespace,
			Subsystem: p.Subsystem,
			Name:      "request_duration_seconds",
			Help:      "The request latencies in seconds.",
		},
	)
	prometheus.MustRegister(p.reqDur)

	p.reqSz = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: p.Namespace,
			Subsystem: p.Subsystem,
			Name:      "request_size_bytes",
			Help:      "The request sizes in bytes.",
		},
	)
	prometheus.MustRegister(p.reqSz)

	p.resSz = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: p.Namespace,
			Subsystem: p.Subsystem,
			Name:      "response_size_bytes",
			Help:      "The response sizes in bytes.",
		},
	)
	prometheus.MustRegister(p.resSz)
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (p *Prometheus) shouldIgnore(handler string) bool {
	p.Ignored.RLock()
	defer p.Ignored.RUnlock()
	if p.Ignored.values == nil {
		return false
	}
	_, found := p.Ignored.values[handler]
	return found
}
