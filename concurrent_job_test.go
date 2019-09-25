package work

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

func Test_ConcurrentJob(t *testing.T) {
	logger := logrus.WithFields(logrus.Fields{
		"method": "Test_ConcurrentJob",
	})
	r := gin.Default()
	health := NewHealthCheck(WithEngine(r), WithHealthPath("my-health"))
	defer health.Close()
	healthAdapter := WithHealthCheck(health, WithErrorPercentage(float64(.70), 1))
	prom := NewPrometheus(WithEngine(r), WithNamespace("go_work_test"), WithSubSystem("Test_ConcurrentJob"), WithMetricsPath("metrics-test-path"))
	srv := &http.Server{
		Addr:    ":8888",
		Handler: r,
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatalf("listen: %s\n", err)
		}
	}()
	metricsHandler := WithPrometheus(prom)
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
	job := Job{
		Args:    Args{},
		Handler: "concurrent-test",
	}
	handler := Adapt(func(j *Job) error {
		atomic.AddUint64(&cnt, 1)
		if logger, ok := j.Ctx.Get(AggregateLogger); ok {
			logger.(Logger).Debugf("current cnt: %d", atomic.LoadUint64(&cnt))
		} else {
			panic(fmt.Sprintf("no logger - %+v ---- %+v", j.Ctx, logger))
		}
		j.Ctx.SetStatus(StatusSuccess)
		return nil
	}, healthAdapter, aggregateHandler, metricsHandler)

	var factory WorkerFactory = NewCommonWorkerFactory

	concurrent, err := NewConcurrentJob(
		job,
		factory,
		PerformEveryWithSync,
		1*time.Second,
		50,
		50*time.Millisecond,
		logger,
	)
	is.NoErr(err)
	err = concurrent.Register("concurrent-test", handler)
	is.NoErr(err)
	go func() {
		// has to be in a goroutine, since it's a blocking syncronise call (normally that's okay but not for this test)
		err := concurrent.Start()
		// intentionally not calling concurrent.Stop() which would clean resources - we need to do things diff for the test
		is.NoErr(err)
	}()

	time.Sleep(5 * time.Second) // sleep, so some concurrent work can be done

	tr := &http.Transport{
		// #nosec G402 - just a test http server
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("http://localhost:8888/metrics-test-path")
	is.NoErr(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	is.NoErr(err)
	t.Log(string(body))
	is.True(resp.StatusCode == 200)

	resp, err = client.Get("http://localhost:8888/my-health")
	is.NoErr(err)
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	is.NoErr(err)
	t.Log("health: "+string(body), " / ", resp.StatusCode)
	is.True(resp.StatusCode == 200)

	cntFinal := atomic.LoadUint64(&cnt)
	t.Log("cnt:", cntFinal)
	p, err := os.FindProcess(os.Getpid())
	is.NoErr(err)
	err = p.Signal(os.Interrupt)
	is.NoErr(err)
	time.Sleep(50 * time.Millisecond) // give the signal a moment to be sent and caught

	t.Log("#############################################################################################################################################")
	health2 := NewHealthCheck(WithEngine(r), WithHealthPath("my-health-bad"), WithHealthTicker(time.NewTicker(1*time.Millisecond)))
	defer health2.Close()
	healthAdapter = WithHealthCheck(health2, WithErrorPercentage(float64(.70), 1))
	badhandler := Adapt(func(j *Job) error {
		j.Ctx.SetStatus(StatusInternalError)
		return nil
	}, healthAdapter, aggregateHandler, metricsHandler)
	is.NoErr(err)
	concurrentTwo, err := NewConcurrentJob(
		job,
		factory,
		PerformEveryWithSync,
		1*time.Second,
		50,
		50*time.Millisecond,
		logger,
	)
	is.NoErr(err)
	err = concurrentTwo.Register("badhandler-test", badhandler)
	concurrentTwo.Job.Handler = "badhandler-test"

	is.NoErr(err)
	go func() {
		// has to be in a goroutine, since it's a blocking syncronise call (normally that's okay but not for this test)
		err := concurrentTwo.Start()
		// intentionally not calling concurrent.Stop() which would clean resources - we need to do things diff for the test
		is.NoErr(err)
	}()
	time.Sleep(2 * time.Second) // sleep, so some concurrent work can be done

	resp, err = client.Get("http://localhost:8888/my-health-bad")
	is.NoErr(err)
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	is.NoErr(err)
	t.Log("health: "+string(body), " / ", resp.StatusCode)
	is.True(resp.StatusCode == 500)
	err = p.Signal(os.Interrupt)
	is.NoErr(err)
	time.Sleep(50 * time.Millisecond) // give the signal a moment to be sent and caught

	t.Log("#############################################################################################################################################")
	health3 := NewHealthCheck(WithEngine(r), WithHealthPath("my-health-bad-num-req"), WithHealthTicker(time.NewTicker(1*time.Millisecond)))
	defer health3.Close()
	healthAdapterByReqCnt := WithHealthCheck(health3, WithErrorPercentageByRequestCount(float64(.70), 1, 4))
	badhandler = Adapt(func(j *Job) error {
		j.Ctx.SetStatus(StatusInternalError)
		return nil
	}, healthAdapterByReqCnt, aggregateHandler, metricsHandler)
	is.NoErr(err)
	concurrentThree, err := NewConcurrentJob(
		job,
		factory,
		PerformEveryWithSync,
		1*time.Second,
		50,
		50*time.Millisecond,
		logger,
	)
	is.NoErr(err)
	err = concurrentThree.Register("badhandler-test-withHealthByReqCnt", badhandler)
	is.NoErr(err)
	concurrentThree.Job.Handler = "badhandler-test-withHealthByReqCnt"
	go func() {
		// has to be in a goroutine, since it's a blocking syncronise call (normally that's okay but not for this test)
		err := concurrentThree.Start()
		// intentionally not calling concurrent.Stop() which would clean resources - we need to do things diff for the test
		is.NoErr(err)
	}()
	time.Sleep(2 * time.Second) // sleep, so some concurrent work can be done

	resp, err = client.Get("http://localhost:8888/my-health-bad-num-req")
	is.NoErr(err)
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	is.NoErr(err)
	t.Log("health: "+string(body), " / ", resp.StatusCode)
	is.True(resp.StatusCode == 500)
	err = p.Signal(os.Interrupt)
	is.NoErr(err)
	time.Sleep(50 * time.Millisecond) // give the signal a moment to be sent and caught

}

func Test_ConcurrentJobPerformReceive(t *testing.T) {
	logger := logrus.WithFields(logrus.Fields{
		"method": "Test_ConcurrentJobPerformReceive",
	})
	r := gin.Default()
	health := NewHealthCheck(WithEngine(r), WithHealthPath("my-health"))
	defer health.Close()
	healthAdapter := WithHealthCheck(health, WithErrorPercentage(float64(.70), 1))
	prom := NewPrometheus(WithEngine(r), WithNamespace("go_work_test"), WithSubSystem("Test_ConcurrentJobPerformReceive"), WithMetricsPath("metrics-test-path"))
	srv := &http.Server{
		Addr:    ":8888",
		Handler: r,
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatalf("listen: %s\n", err)
		}
	}()
	metricsHandler := WithPrometheus(prom)
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
	job := Job{
		Args:    Args{},
		Handler: "concurrent-test",
	}

	ch := make(chan string, 1000)
	for i := 0; i < 1000; i++ {
		ch <- fmt.Sprintf("message: %d", i)
	}
	close(ch)

	handler := Adapt(func(j *Job) error {
		logger, _ := j.Ctx.Get(AggregateLogger)
		if d, ok := j.Args["receivedData"]; ok {
			logger.(Logger).Infof("current data: %v", d)
			j.Ctx.SetStatus(StatusSuccess)
		}
		atomic.AddUint64(&cnt, 1)
		logger.(Logger).Debugf("current cnt: %d", atomic.LoadUint64(&cnt))
		j.Ctx.SetStatus(StatusSuccess)
		return nil
	}, healthAdapter, aggregateHandler, metricsHandler)

	var factory WorkerFactory = NewCommonWorkerFactory

	concurrent, err := NewConcurrentJob(
		job,
		factory,
		PerformReceiveWithSync,
		0,
		50,
		1*time.Millisecond,
		logger,
		WithChannel(ch),
	)
	is.NoErr(err)
	err = concurrent.Register("concurrent-test", handler)
	is.NoErr(err)
	go func() {
		// has to be in a goroutine, since it's a blocking syncronise call (normally that's okay but not for this test)
		err := concurrent.Start()
		// intentionally not calling concurrent.Stop() which would clean resources - we need to do things diff for the test
		is.NoErr(err)
	}()

	time.Sleep(5 * time.Second) // sleep, so some concurrent work can be done

	tr := &http.Transport{
		// #nosec G402 - just a test http server
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("http://localhost:8888/metrics-test-path")
	is.NoErr(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	is.NoErr(err)
	t.Log(string(body))
	is.True(resp.StatusCode == 200)

	resp, err = client.Get("http://localhost:8888/my-health")
	is.NoErr(err)
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	is.NoErr(err)
	t.Log("health: "+string(body), " / ", resp.StatusCode)
	is.True(resp.StatusCode == 200)

	cntFinal := atomic.LoadUint64(&cnt)
	t.Log("cnt:", cntFinal)
	p, err := os.FindProcess(os.Getpid())
	is.NoErr(err)
	err = p.Signal(os.Interrupt)
	is.NoErr(err)
	time.Sleep(50 * time.Millisecond) // give the signal a moment to be sent and caught
}
