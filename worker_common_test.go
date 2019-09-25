package work

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

func Test_CommonWorker_Perform(t *testing.T) {
	logger := logrus.WithFields(logrus.Fields{
		"method": "Test_CommonWorker_Perform",
	})
	useBanner := true
	useUTC := true
	aggregateHandler := WithAggregateLogger(
		useBanner,
		time.RFC3339,
		useUTC,
		"requestID",
		[]byte("RequestID"),
		WithBanner(true))
	var hit bool
	wg := &sync.WaitGroup{}
	is := is.New(t)
	w := NewCommonWorker(logger)
	defer func() { _ = w.Stop() }() // reclaim resources
	wg.Add(1)
	err := w.Register(
		"async",
		Adapt(func(j *Job) error {
			logger, _ := j.Ctx.Get(AggregateLogger)
			logger.(Logger).Info("async job")
			hit = true
			wg.Done()
			j.Ctx.SetStatus(StatusSuccess)
			return nil
		}, aggregateHandler))
	is.NoErr(err)
	err = w.Perform(&Job{
		Handler: "async",
	})
	is.NoErr(err)
	wg.Wait()
	is.True(hit)

	// now let's do WithSync(true)
	hit = false
	err = w.Register(
		"sync",
		Adapt(func(j *Job) error {
			logger, _ := j.Ctx.Get(AggregateLogger)
			logger.(Logger).Info("sync job")
			hit = true
			j.Ctx.SetStatus(StatusSuccess)
			return nil
		}, aggregateHandler))
	is.NoErr(err)
	err = w.Perform(
		&Job{Handler: "sync"},
		WithSync(true))
	is.NoErr(err)

	// test timeouts
	hit = false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w2 := NewCommonWorkerWithContext(ctx, logger)
	defer func() { _ = w2.Stop() }() // reclaim resources
	err = w2.Register(
		"timed-out",
		Adapt(func(j *Job) error {
			logger, _ := j.Ctx.Get(AggregateLogger)
			deadline, ok := j.Ctx.Deadline()
			logger.(Logger).Debugf("Deadline: %v / %v / now == %v", deadline, ok, time.Now())
			select {
			case <-j.Ctx.Done():
				logger.(Logger).Debug("timeout")
				j.Ctx.SetStatus(StatusTimeout)
				return nil
			case <-time.After(1 * time.Second):
				logger.(Logger).Debugf("Deadline: %v / %v / now == %v", deadline, ok, time.Now())
				hit = true
				j.Ctx.SetStatus(StatusInternalError)
				logger.(Logger).Error("hit success - but should have timed out")
			}
			return nil
		}, aggregateHandler))
	is.NoErr(err)
	job := Job{
		Handler: "timed-out",
		Timeout: 2 * time.Millisecond,
	}
	err = w2.Perform(&job, WithSync(true))
	is.NoErr(err)
	t.Log("status: ", job.Ctx.Status())
	is.True(job.Ctx.Status() == StatusTimeout)
}

func Test_CommonWorker_PerformEvery(t *testing.T) {
	logger := logrus.WithFields(logrus.Fields{
		"method": "Test_CommonWorker_PerformEvery",
	})
	useBanner := true
	useUTC := true
	aggregateHandler := WithAggregateLogger(
		useBanner,
		time.RFC3339,
		useUTC,
		"requestID",
		[]byte("RequestID"),
		WithBanner(true))

	var cnt uint64
	is := is.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := NewCommonWorkerWithContext(ctx, logger)

	err := w.Register(
		"sync",
		Adapt(func(j *Job) error {
			logger, _ := j.Ctx.Get(AggregateLogger)
			atomic.AddUint64(&cnt, 1)
			logger.(Logger).Debugf("current cnt: %d", atomic.LoadUint64(&cnt))
			j.Ctx.SetStatus(StatusSuccess)
			return nil
		}, aggregateHandler))
	is.NoErr(err)
	t.Log(w)
	go func() {
		err := w.PerformEvery(
			&Job{Handler: "sync"},
			25*time.Millisecond,
			WithSync(true))
		is.NoErr(err)
	}()
	time.Sleep(250 * time.Millisecond) // let it Perform a few times
	cancel()                           // cancel PerformEvery()
	err = w.Stop()                     // now just stop() and wait for every thing to finish
	is.NoErr(err)
	cntFinal := atomic.LoadUint64(&cnt)
	t.Log("cnt:", cntFinal)
	time.Sleep(250 * time.Millisecond) // let's see if it's still Performing in the background.. or if it Stop()/cancel() like requested
	is.Equal(cntFinal, atomic.LoadUint64(&cnt))
}

func Test_CommonWorker_PerformReceive(t *testing.T) {
	logger := logrus.WithFields(logrus.Fields{
		"method": "Test_CommonWorker_PerformReceive",
	})
	useBanner := true
	useUTC := true
	aggregateHandler := WithAggregateLogger(
		useBanner,
		time.RFC3339,
		useUTC,
		"requestID",
		[]byte("RequestID"),
		WithBanner(true))

	var cnt uint64
	is := is.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := NewCommonWorkerWithContext(ctx, logger)

	err := w.Register(
		"sync",
		Adapt(func(j *Job) error {
			logger, _ := j.Ctx.Get(AggregateLogger)
			if d, ok := j.Args["receivedData"]; ok {
				logger.(Logger).Infof("current data: ", d)
				j.Ctx.SetStatus(StatusSuccess)
			}

			atomic.AddUint64(&cnt, 1)
			logger.(Logger).Debugf("current cnt: %d", atomic.LoadUint64(&cnt))
			j.Ctx.SetStatus(StatusSuccess)
			return nil
		}, aggregateHandler))
	is.NoErr(err)
	t.Log(w)
	c := make(chan string, 10)
	c <- "hi mom"
	c <- "hi dad"
	close(c)
	go func() {
		err := w.PerformReceive(
			&Job{Handler: "sync"},
			c,
			WithSync(false))
		is.NoErr(err)
	}()
	time.Sleep(250 * time.Millisecond) // let it Perform a few times
	cancel()                           // cancel PerformEvery()
	err = w.Stop()                     // now just stop() and wait for every thing to finish
	is.NoErr(err)
	cntFinal := atomic.LoadUint64(&cnt)
	t.Log("cnt:", cntFinal)
	time.Sleep(250 * time.Millisecond) // let's see if it's still Performing in the background.. or if it Stop()/cancel() like requested
	is.Equal(cntFinal, atomic.LoadUint64(&cnt))
}
