package work

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/xerrors"
)

// CommonWorker defines the typical common worker
type CommonWorker struct {

	// ctx for the worker (which will cascade down to the Jobs a worker performs)
	ctx context.Context

	// cancel the worker (which will cascade down to the Jobs a worker performs)
	cancel context.CancelFunc

	// handlers registered with the worker
	handlers map[string]Handler

	// moot keeps everything sycronized
	moot *sync.Mutex

	// Logger for the worker
	Logger Logger

	// wg is used to coordinate shutdown and wait for running Jobs
	wg sync.WaitGroup
}

var _ Worker = &CommonWorker{}

// NewCommonWorker creates a new CommonWorker
func NewCommonWorker(l Logger) *CommonWorker {
	return NewCommonWorkerWithContext(context.Background(), l)
}

// NewCommonWorkerWithContext creates a new CommonWorker
func NewCommonWorkerWithContext(ctx context.Context, l Logger) *CommonWorker {
	ctx, cancel := context.WithCancel(ctx)

	return &CommonWorker{
		Logger:   l,
		ctx:      ctx,
		cancel:   cancel,
		handlers: map[string]Handler{},
		moot:     &sync.Mutex{},
	}
}

// NewCommonWorkerFactory is a simple adapter that allows the factory to conform to the WorkerFactory type
func NewCommonWorkerFactory(ctx context.Context, l Logger) Worker {
	var w Worker = NewCommonWorkerWithContext(ctx, l)
	return w
}

// Stop the worker and you must call Stop() to clean up the CommonWorker internal Ctx (or it will leak memory)
func (w *CommonWorker) Stop() error {
	w.Logger.Info("CommonWorker.Stop: stoppping...")
	w.cancel()
	w.wg.Wait()
	w.Logger.Info("CommonWorker.Stop: we're done.")
	return nil
}

// Start the worker
func (w *CommonWorker) Start(ctx context.Context) error {
	w.Logger.Info("CommonWorker.Start: we've begun...")
	return nil
}

// GetContext from the worker
func (w *CommonWorker) GetContext() context.Context {
	return w.ctx
}

// SetContext for the worker
func (w *CommonWorker) SetContext(ctx context.Context) {
	w.ctx = ctx
}

// Perform executes the job.  If WithSync(true) Option then it's a blocking call, the default is false (so async)
func (w *CommonWorker) Perform(job *Job, opt ...Option) error {
	opts := GetOpts(opt...)
	withSync := opts[optionWithSync].(bool)

	// w.Logger.Debugf("Perform: job %s", job)
	if job.Handler == "" {
		err := xerrors.Errorf("CommonWorker.Perform: no handler name given for %s", job)
		w.Logger.Error(err)
		return err
	}
	w.moot.Lock()
	defer w.moot.Unlock()
	if h, ok := w.handlers[job.Handler]; ok {
		if withSync {
			var syncCtx context.Context
			var syncCancel context.CancelFunc
			if job.Timeout != 0 {
				syncCtx, syncCancel = context.WithTimeout(w.ctx, job.Timeout)
			} else {
				syncCtx, syncCancel = context.WithCancel(w.ctx)
			}
			c := NewContext(syncCtx)
			job.Ctx = &c
			err := RunE(func() error { return h(job) }, WithJob(job))
			syncCancel() // no reason to defer this
			if err != nil {
				w.Logger.Error(err)
				return err
			}
			return nil
		} else {
			go func() {
				w.wg.Add(1)
				defer w.wg.Done()
				// goroutine needs it's own ctx
				var jobCtx context.Context
				var jobCancel context.CancelFunc
				if job.Timeout != 0 {
					jobCtx, jobCancel = context.WithTimeout(w.ctx, job.Timeout)
				} else {
					jobCtx, jobCancel = context.WithCancel(w.ctx)
				}
				defer jobCancel()
				c := NewContext(jobCtx)
				job.Ctx = &c
				err := RunE(func() error { return h(job) }, WithJob(job))
				if err != nil {
					w.Logger.Error(err)
				}
				w.Logger.Debugf("CommonWorker.Perform: completed job %s", job)
			}()
			return nil
		}
	}
	err := xerrors.Errorf("CommonWorker.Perform: no handler mapped for name %s", job.Handler)
	w.Logger.Error(err)
	return err
}

// PerformEvery executes the job on the interval.
// if WithSync(true) Option, then the operation blocks until it's done which means only one instance can be executed at a time.
// the default is WithSync(false), so there's not blocking and you could get multiple instances running at a time if the latency is longer than the interval.
func (w *CommonWorker) PerformEvery(job *Job, interval time.Duration, opt ...Option) error {
	if interval == 0 {
		return xerrors.New("CommonWorker.PerformEvery: time.Duration cannot be 0")
	}
	// get the first execution in before waiting the interval
	// passing in opt... it could include options like: WithSync(bool)
	if err := w.Perform(job, opt...); err != nil {
		w.Logger.Errorf("CommonWorker.PerformEvery: first job error == %s", err.Error())
	}
	tick := time.Tick(interval)
	for {
		select {
		case <-tick:
			// passing in opt... it could include options like: WithSync(bool)
			if err := w.Perform(job, opt...); err != nil {
				w.Logger.Errorf("CommonWorker.PerformEvery: tick job error == %s", err.Error())
			}
		case <-w.ctx.Done():
			workerNumber := int64(-1)
			if num, ok := job.Ctx.Value("workerNumber").(int64); ok {
				workerNumber = num
			}
			w.Logger.Infof("CommonWorker.PerformEvery: stopping with context %v  - worker %d", w.ctx.Err(), workerNumber)
			w.cancel()
			return nil
		}
	}
}

// PerformReceive will loop on receiving data from the readChan (passed as a simple interface{}).
// This uses the Job ctx for timeouts and cancellation (just like the rest of the framework)
func (w *CommonWorker) PerformReceive(job *Job, readChan interface{}, opt ...Option) error {
	ch, err := WrapChannel(readChan)
	if err != nil {
		return fmt.Errorf("PerformReceive: unable to wrap channel - %s", err.Error())
	}
	cnt := 0
	for {
		cnt += 1
		select {
		case data, ok := <-ch:
			if !ok { // handle the close chan not blocking properly
				w.cancel()
				return nil
			}
			jobCopy := job.Copy()
			jobCopy.Args["receivedData"] = data
			// passing in opt... it could include options like: WithSync(bool)
			if err := w.Perform(&jobCopy, opt...); err != nil {
				w.Logger.Errorf("CommonWorker.PerformReceive: job error == %s", err.Error())
			}
		case <-w.ctx.Done():
			if job.Ctx != nil {
				workerNumber := int64(-1)
				if num, ok := job.Ctx.Value("workerNumber").(int64); ok {
					workerNumber = num
				}
				w.Logger.Infof("CommonWorker.PerformReceive: stopping with context %v  - worker %d", w.ctx.Err(), workerNumber)
			}
			w.cancel()
			return nil
		}
	}
	// no need for return here, since it would be unreachable.
}

// PerformIn performs a job after the "in" time has expired.
func (w *CommonWorker) PerformIn(job *Job, d time.Duration) error {
	w.Logger.Debugf("CommonWorker.PerformIn: scheduling job %s for %s", job.Handler, d)
	if d == 0 {
		return xerrors.New("CommonWorker.PerformIn: time.Duration cannot be 0")
	}
	go func() {
		w.wg.Add(1)
		defer w.wg.Done()
		select {
		case <-time.After(d):
			if err := w.Perform(job); err != nil {
				w.Logger.Errorf("CommonWorker.PerformIn: after job error == %s", err.Error())
			}
		case <-w.ctx.Done():
			w.cancel()
		}
	}()
	return nil
}

// PerformAt performs a job at a particular time using a goroutine.
func (w *CommonWorker) PerformAt(job *Job, t time.Time) error {
	if (time.Time{}) == t {
		return xerrors.New("CommonWorker.PerformAt: time.Time cannot be 0")
	}
	return w.PerformIn(job, time.Until(t))
}

// Register Handler with the worker
func (w *CommonWorker) Register(name string, h Handler) error {
	w.moot.Lock()
	defer w.moot.Unlock()
	if _, ok := w.handlers[name]; ok {
		return xerrors.Errorf("CommonWorker.Register: handler already mapped for name %s", name)
	}
	w.handlers[name] = h
	return nil
}
