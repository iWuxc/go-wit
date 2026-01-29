package queue

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/queue/contract"
)

// ResultWriter is a client interface to write result data for a task.
// It writes the data to the redis instance the server is connected to.
type ResultWriter struct {
	ID     string // task ID this writer is responsible for
	QName  string // queue name the task belongs to
	Broker contract.BrokerInterface
	Ctx    context.Context // context associated with the task
}

// Write writes the given data as a result of the task the ResultWriter is associated with.
func (w *ResultWriter) Write(data []byte) (n int, err error) {
	select {
	case <-w.Ctx.Done():
		return 0, fmt.Errorf("failed to result task result: %v", w.Ctx.Err())
	default:
	}
	return w.Broker.WriteResult(w.QName, w.ID, data)
}

// TaskID returns the ID of the task the ResultWriter is associated with.
func (w *ResultWriter) TaskID() string {
	return w.ID
}
