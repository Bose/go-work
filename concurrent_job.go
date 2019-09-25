package work

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/xerrors"
)

type PerformWith int

const (
	PerformWithUnknown PerformWith = iota
	PerformWithSync
	PerformWithAsync
	PerformEveryWithSync
	PerformEveryWithAsync
	PerformReceiveWithSync
	PerformReceiveWithAsync
)

func (p PerformWith) String() string {
	names := [...]string{
		"PerformWithUnknown",
		"PerformWithWithSync",
		"PerformWithWithAsync",
		"PerformWithEveryWithSync",
		"PerformWithEveryWithAsync",
	}
	if p < PerformWithUnknown || p > PerformEveryWithAsync {
		return names[PerformWithUnknown]
	}
	return names[p]
}

type WorkerFactory func(context.Context, Logger) Worker

// ConcurrentJob represents a job to be run concurrently
type ConcurrentJob struct {
	// workerFactory used to make workers for the job
	workerFactory WorkerFactory

	// PerformWith if the job will be performed with PerformWithEveryWithSync as a reoccuring job
	// or just once as PerformWithWithSync
	PerformWith PerformWith

	// PerformEvery defines the duration between executions of PerformWithEveryWithSync jobs
	PerformEvery time.Duration

	// MaxWorkers for the concurrent Job
	MaxWorkers int64

	// Job to run concurrently
	Job Job

	// hanlders that have been registered
	handlers map[string]Handler

	// runningWorkers is the current count of what's running
	runningWorkers int64

	// startInterval defines duration between Job starts as we ramp up
	startInterval time.Duration

	// cts is the context for all the work
	ctx context.Context

	// cancel allows all the work to be cancelled
	cancel context.CancelFunc

	// logger is an interface to the work's logger
	logger Logger

	// moot allows things to be syncronized safely
	moot *sync.Mutex

	// wg is used to coordinate shutdown and wait for running work
	wg sync.WaitGroup

	PerformReceiveChan interface{}
}

// NewConcurrentJob makes a new job
func NewConcurrentJob(
	job Job,
	workerFactory WorkerFactory,
	performWith PerformWith,
	performEvery time.Duration,
	maxWorker int64,
	startInterval time.Duration,
	logger Logger,
	opt ...Option,
) (ConcurrentJob, error) {
	if maxWorker < 1 {
		return ConcurrentJob{}, xerrors.New("NewConcurrentJob: maxWorker must be more than 0")
	}
	if performWith == PerformEveryWithSync && performEvery == 0 {
		return ConcurrentJob{}, xerrors.New("NewConcurrentJob: you must specify an interval to PerformEveryWithSync")
	}
	opts := GetOpts(opt...)
	ch, _ := opts.Get(optionWithChannel)
	if performWith == PerformReceiveWithSync && ch == nil {
		return ConcurrentJob{}, xerrors.New("NewConcurrentJob: you must specify a WithChannel() option to PeformReceiveWithSync")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return ConcurrentJob{
		Job:                job,
		workerFactory:      workerFactory,
		PerformWith:        performWith,
		PerformEvery:       performEvery,
		MaxWorkers:         maxWorker,
		startInterval:      startInterval,
		cancel:             cancel,
		ctx:                ctx,
		moot:               &sync.Mutex{},
		handlers:           map[string]Handler{},
		logger:             logger,
		PerformReceiveChan: ch,
	}, nil
}

// Register a handler for the Job's workers
func (j *ConcurrentJob) Register(name string, h Handler) error {
	j.moot.Lock()
	defer j.moot.Unlock()
	if _, ok := j.handlers[name]; ok {
		return xerrors.Errorf("ConcurrentJob.Register: handler already mapped for name %s", name)
	}
	j.handlers[name] = h
	return nil
}
func (j *ConcurrentJob) workersIncrement() {
	atomic.AddInt64(&j.runningWorkers, 1)
}
func (j *ConcurrentJob) workersDecrement() {
	atomic.AddInt64(&j.runningWorkers, -1)
}
func (j *ConcurrentJob) RunningWorkers() int64 {
	return atomic.LoadInt64(&j.runningWorkers)
}

// addWorker adds a new worker to the runningWorkers performs the job
func (j *ConcurrentJob) addWorker() {
	if j.RunningWorkers() >= j.MaxWorkers {
		// j.logger.Debugf("ConcurrentJob.addWorker: already running %d, so no op", j.MaxWorkers)
		return
	}
	var w Worker = j.workerFactory(j.ctx, j.logger)
	wCtx := w.GetContext()
	w.SetContext(context.WithValue(wCtx, "workerNumber", j.RunningWorkers()+1))

	if j.PerformWith == PerformWithAsync || j.PerformWith == PerformEveryWithAsync {
		j.logger.Errorf("ConcurrentJob.addWorker: %s not a support Perform method", j.PerformWith)
	}
	if err := w.Register(j.Job.Handler, j.handlers[j.Job.Handler]); err != nil {
		j.logger.Error("ConcurrentJob.addWorker: error registring handler == %s", err.Error())
	}
	if j.PerformWith == PerformWithSync {
		j.workersIncrement()
		go func() {
			defer j.workersDecrement()
			j.wg.Add(1)
			defer j.wg.Done()
			jobCopy := j.Job.Copy() // you MUST copy this before Performing the job
			if err := w.Perform(&jobCopy, WithSync(true)); err != nil {
				j.logger.Errorf("ConcurrentJob.addWorker: Perform error %s", err.Error())
			}
		}()
	}
	if j.PerformWith == PerformEveryWithSync {
		j.workersIncrement()
		go func() {
			defer j.workersDecrement()
			j.wg.Add(1)
			defer j.wg.Done()
			jobCopy := j.Job.Copy() // you MUST copy this before Performing the job
			if err := w.PerformEvery(&jobCopy, j.PerformEvery, WithSync(true)); err != nil {
				j.logger.Errorf("ConcurrentJob.addWorker: PerformEvery error %s", err.Error())
			}
		}()
	}
	if j.PerformWith == PerformReceiveWithSync {
		j.workersIncrement()
		go func() {
			defer j.workersDecrement()
			j.wg.Add(1)
			defer j.wg.Done()
			jobCopy := j.Job.Copy() // you MUST copy this before Performing the job
			if err := w.PerformReceive(&jobCopy, j.PerformReceiveChan, WithSync(true)); err != nil {
				j.logger.Errorf("ConcurrentJob.addWorker: PerformReceive error %s", err.Error())
			}
		}()

	}
}

// Stop all the work and you must call Stop() to clean up the ConcurrentJob Ctx (or it will leak memory)
func (j *ConcurrentJob) Stop() {
	j.logger.Info("ConcurrentJob.Stop: stoppping...")
	j.cancel()
	j.wg.Wait()
	j.ctx.Done()
	j.logger.Info("ConcurrentJob.Stop: we're done.")
}

// Start all the work
func (j *ConcurrentJob) Start() error {
	if _, exists := j.handlers[j.Job.Handler]; !exists {
		return xerrors.New(fmt.Sprintf("ConcurrentJob.Start: Job.Handler %s is not registered", j.Job.Handler))
	}
	signalCh := make(chan os.Signal, 1024)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, os.Interrupt)

	go func() {
		j.addWorker()
		j.runJanitor(j.ctx)
	}()
	for {
		select {
		case sig := <-signalCh:
			j.logger.Infof("Concurrent.Job: %v signal received", sig)
			switch sig {
			case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL:
				// Return any errors during shutdown.
				j.Stop()
				return nil
			default:
				// Return any errors during shutdown.
				j.Stop()
				return nil
			}
		}
	}
}

// runJanitor keeps everything running smoothly and keeps the right number of workers working
func (j *ConcurrentJob) runJanitor(ctx context.Context) {
	tick := time.Tick(j.startInterval)
	for {
		select {
		case <-tick:
			for i := j.RunningWorkers(); i <= j.MaxWorkers; i++ {
				j.addWorker()
			}
		case <-ctx.Done():
			j.logger.Debug("ConcurrentJob.runJanitor: we are done.")
			return
		}
	}
}
