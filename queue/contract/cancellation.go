package contract

import (
	"context"
	"sync"
)

// Cancellations is a collection that holds cancel functions for all active tasks.
//
// Cancellations are safe for concurrent use by multiple goroutines.
type Cancellations struct {
	mu         sync.Mutex
	cancelFunc map[string]context.CancelFunc
}

// NewCancellations returns a Cancellations instance.
func NewCancellations() *Cancellations {
	return &Cancellations{
		cancelFunc: make(map[string]context.CancelFunc),
	}
}

// Add adds a new cancel func to the collection.
func (c *Cancellations) Add(id string, fn context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelFunc[id] = fn
}

// Delete deletes a cancel func from the collection given an id.
func (c *Cancellations) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cancelFunc, id)
}

// Get returns a cancel func given an id.
func (c *Cancellations) Get(id string) (fn context.CancelFunc, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn, ok = c.cancelFunc[id]
	return fn, ok
}
