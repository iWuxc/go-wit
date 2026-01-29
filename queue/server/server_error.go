package server

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/queue"
)

// ErrServerClosed indicates that the operation is now illegal because of the server has been shutdown.
var ErrServerClosed = errors.New("go-kit: Server closed")

// An ErrorHandler handles an error occurred during task processing.
type ErrorHandler interface {
	HandleError(ctx context.Context, task *queue.Task, err error)
}

// The ErrorHandlerFunc type is an adapter to allow the use of  ordinary functions as a ErrorHandler.
// If f is a function with the appropriate signature, ErrorHandlerFunc(f) is a ErrorHandler that calls f.
type ErrorHandlerFunc func(ctx context.Context, task *queue.Task, err error)

// HandleError calls fn(ctx, task, err)
func (fn ErrorHandlerFunc) HandleError(ctx context.Context, task *queue.Task, err error) {
	fn(ctx, task, err)
}
