package work

import (
	"context"
	"sync"
)

// Context for the Job and is reset for every execution
type Context struct {
	// context.Context allows it to be a compatible context
	context.Context

	// status records the job's execution Status (StatusSuccess, StatusBadRequest, etc)
	status Status

	// moot keeps things sycronized
	moot *sync.Mutex

	// keys stores key/values for the context
	keys map[string]interface{}
}

// NewContext factory
func NewContext(ctx context.Context) Context {
	return Context{
		ctx,
		-1,
		&sync.Mutex{},
		map[string]interface{}{},
	}
}

// Get a key/value
func (c *Context) Get(k string) (interface{}, bool) {
	c.moot.Lock()
	defer c.moot.Unlock()
	value, exists := c.keys[k]
	return value, exists
}

// Set a key/value
func (c *Context) Set(k string, v interface{}) {
	c.moot.Lock()
	defer c.moot.Unlock()
	c.keys[k] = v
}

// Status retrieves the context's status
func (c *Context) Status() Status {
	return c.status
}

// SetStatus for the context
func (c *Context) SetStatus(s Status) {
	if c.status != 0 {

	}
	c.moot.Lock()
	defer c.moot.Unlock()
	c.status = s
}
