package errors

import "fmt"

var (
	// ErrNoProcessableTask indicates that there are no tasks ready to be processed.
	ErrNoProcessableTask = New("no tasks are ready for processing")

	// ErrDuplicateTask indicates that another task with the same unique key holds the uniqueness lock.
	ErrDuplicateTask = New("task already exists")

	// ErrTaskIdConflict indicates that another task with the same task ID already exist
	ErrTaskIdConflict = New("task id conflicts with another task")
)

// QueueNotFoundError indicates that a queue with the given name does not exist.
type QueueNotFoundError struct {
	Queue string // queue name
}

func (e *QueueNotFoundError) Error() string {
	return fmt.Sprintf("queue %q does not exist", e.Queue)
}

// IsQueueNotFound reports whether any error in error's chain is of type QueueNotFoundError.
func IsQueueNotFound(err error) bool {
	var target *QueueNotFoundError
	return As(err, &target)
}

// QueueNotEmptyError indicates that the given queue is not empty.
type QueueNotEmptyError struct {
	Queue string // queue name
}

func (e *QueueNotEmptyError) Error() string {
	return fmt.Sprintf("queue %q is not empty", e.Queue)
}

// IsQueueNotEmpty reports whether any error in error's chain is of type QueueNotEmptyError.
func IsQueueNotEmpty(err error) bool {
	var target *QueueNotEmptyError
	return As(err, &target)
}
