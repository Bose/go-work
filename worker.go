package work

import (
	"context"
	"time"
)

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
