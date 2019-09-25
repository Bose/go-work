package work

import (
	"time"
)

// Job to be processed by a Worker
type Job struct {
	// Queue the job should be placed into
	Queue string

	// ctx related to the execution of a job - Perform(job) gets a new ctx everytime
	Ctx *Context

	// Args that will be passed to the Handler as the 2nd parameter when run
	Args Args

	// Handler that will be run by the worker
	Handler string

	// Timeout for every execution of job
	Timeout time.Duration
}

// Copy provides a deep copy of the Job
func (j *Job) Copy() Job {
	newJob := Job{
		Queue:   j.Queue,
		Ctx:     nil,
		Args:    Args{},
		Handler: j.Handler,
		Timeout: j.Timeout,
	}
	for k, v := range j.Args {
		newJob.Args[k] = v
	}
	return newJob
}
