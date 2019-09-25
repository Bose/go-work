package work

import (
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

func Test_WithAggregateLogger(t *testing.T) {
	logger := logrus.WithFields(logrus.Fields{
		"method": "Test_WithGinLogrus",
	})
	var hit bool
	wg := &sync.WaitGroup{}
	is := is.New(t)
	w := NewCommonWorker(logger)
	wg.Add(1)

	useBanner := true
	useUTC := true

	tracer, reporter, closer, err := InitTracing("with-aggregater-logger-example::test-hostname", "localhost:5775")
	if err != nil {
		panic("unable to init tracing")
	}
	defer closer.Close()
	defer reporter.Close()
	opentracing.SetGlobalTracer(tracer)

	traceHandler := WithOpenTracing([]byte("api-request-"))

	aggregateHandler := WithAggregateLogger(
	//	logger,
		useBanner,
		time.RFC3339,
		useUTC,
		"requestID",
		[]byte("RequestID"),
		WithBanner(true))

	jobHandler := func(j *Job) error {
		logger, _ := j.Ctx.Get(AggregateLogger)
		logger.(Logger).Debugf("Hi from inside job")
		hit = true
		wg.Done()
		j.Ctx.SetStatus(StatusSuccess)
		t, _ := j.Ctx.Get("tracing-context")
		logger.(Logger).Debugf("tracing-context: %v", t)
		return nil
	}
	err = w.Register("Test_WithAggregateLogger", Adapt(jobHandler, traceHandler, aggregateHandler))
	is.NoErr(err)
	err = w.Perform(&Job{
		Handler: "Test_WithAggregateLogger",
	})
	is.NoErr(err)
	wg.Wait()
	is.True(hit)
}
