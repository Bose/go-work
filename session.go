package work

import (
	"sync"
)

// Session that is used to pass session info to a Job
// this is a good spot to put things like *redis.Pool or *sqlx.DB for outbox connection pools
type Session struct {
	moot *sync.Mutex
	// keys stores key/values for the Session
	keys map[string]interface{}
}

// NewSession factory
func NewSession() Session {
	return Session{
		&sync.Mutex{},
		map[string]interface{}{},
	}
}

// Get a key/value
func (c *Session) Get(k string) (interface{}, bool) {
	c.moot.Lock()
	defer c.moot.Unlock()
	value, exists := c.keys[k]
	return value, exists
}

// Set a key/value
func (c *Session) Set(k string, v interface{}) {
	c.moot.Lock()
	defer c.moot.Unlock()
	c.keys[k] = v
}