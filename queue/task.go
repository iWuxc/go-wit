package queue

import (
	"fmt"
	"github.com/iWuxc/go-wit/queue/contract"
	"time"
)

// Task represents a unit of work to be performed.
type Task struct {
	// typename indicates the type of task to be performed.
	typename string

	// payload holds data needed to perform the task.
	payload []byte

	// opts holds options for the task.
	Opts []OptionInterface

	// w is the ResultWriter for the task.
	w *ResultWriter
}

func (t *Task) Type() string    { return t.typename }
func (t *Task) Payload() []byte { return t.payload }

// ResultWriter returns a pointer to the ResultWriter associated with the task.
//
// Nil pointer is returned if called on a newly created task (i.e. task created by calling NewTask).
// Only the tasks passed to Handler.ProcessTask have a valid ResultWriter pointer.
func (t *Task) ResultWriter() *ResultWriter { return t.w }

// NewTask returns a new Task given a type name and payload data.
// Options can be passed to configure task processing behavior.
func NewTask(typename string, payload []byte, opts ...OptionInterface) *Task {
	return &Task{
		typename: typename,
		payload:  payload,
		Opts:     opts,
	}
}

// NewTaskWithWriter creates a task with the given typename, payload and ResultWriter.
func NewTaskWithWriter(typename string, payload []byte, w *ResultWriter) *Task {
	return &Task{
		typename: typename,
		payload:  payload,
		w:        w,
	}
}

// A TaskInfo describes a task and its metadata.
type TaskInfo struct {
	// ID is the identifier of the task.
	ID string

	// Queue is the name of the queue in which the task belongs.
	Queue string

	// Type is the type name of the task.
	Type string

	// Payload is the payload data of the task.
	Payload []byte

	// State indicates the task state.
	State TaskState

	// MaxRetry is the maximum number of times the task can be retried.
	MaxRetry int

	// Retried is the number of times the task has retried so far.
	Retried int

	// LastErr is the error message from the last failure.
	LastErr string

	// LastFailedAt is the time of the last failure if any.
	// If the task has no failures, LastFailedAt is zero time (i.e. time.Time{}).
	LastFailedAt time.Time

	// Timeout is the duration the task can be processed by Handler before being retried,
	// zero if not specified
	Timeout time.Duration

	// Deadline is the deadline for the task, zero value if not specified.
	Deadline time.Time

	// NextProcessAt is the time the task is scheduled to be processed,
	// zero if not applicable.
	NextProcessAt time.Time

	// IsOrphaned describes whether the task is left in active state with no worker processing it.
	// An orphaned task indicates that the worker has crashed or experienced network failures and was not able to
	// extend its lease on the task.
	//
	// This task will be recovered by running a server against the queue the task is in.
	// This field is only applicable to tasks with TaskStateActive.
	IsOrphaned bool

	// Retention is duration of the retention period after the task is successfully processed.
	Retention time.Duration

	// CompletedAt is the time when the task is processed successfully.
	// Zero value (i.e. time.Time{}) indicates no value.
	CompletedAt time.Time

	// Result holds the result data associated with the task.
	// Use ResultWriter to write result data from the Handler.
	Result []byte
}

// If t is non-zero, returns time converted from t as unix time in seconds.
// If t is zero, returns zero value of time.Time.
func fromUnixTimeOrZero(t int64) time.Time {
	if t == 0 {
		return time.Time{}
	}
	return time.Unix(t, 0)
}

func NewTaskInfo(msg *contract.Task, state contract.TaskState, nextProcessAt time.Time, result []byte) *TaskInfo {
	info := TaskInfo{
		ID:            msg.ID,
		Queue:         msg.Queue,
		Type:          msg.Type,
		Payload:       msg.Payload,
		MaxRetry:      msg.Retry,
		Retried:       msg.Retried,
		LastErr:       msg.ErrorMsg,
		Timeout:       time.Duration(msg.Timeout) * time.Second,
		Deadline:      fromUnixTimeOrZero(msg.Deadline),
		Retention:     time.Duration(msg.Retention) * time.Second,
		NextProcessAt: nextProcessAt,
		LastFailedAt:  fromUnixTimeOrZero(msg.LastFailedAt),
		CompletedAt:   fromUnixTimeOrZero(msg.CompletedAt),
		Result:        result,
	}

	switch state {
	case contract.TaskStateActive:
		info.State = TaskStateActive
	case contract.TaskStatePending:
		info.State = TaskStatePending
	case contract.TaskStateScheduled:
		info.State = TaskStateScheduled
	case contract.TaskStateRetry:
		info.State = TaskStateRetry
	case contract.TaskStateArchived:
		info.State = TaskStateArchived
	case contract.TaskStateCompleted:
		info.State = TaskStateCompleted
	default:
		panic(fmt.Sprintf("internal error: unknown state: %d", state))
	}
	return &info
}

// TaskState denotes the state of a task.
type TaskState int

const (
	// TaskStateActive Indicates that the task is currently being processed by Handler.
	TaskStateActive TaskState = iota + 1

	// TaskStatePending Indicates that the task is ready to be processed by Handler.
	TaskStatePending

	// TaskStateScheduled Indicates that the task is scheduled to be processed some time in the future.
	TaskStateScheduled

	// TaskStateRetry Indicates that the task has previously failed and scheduled to be processed some time in the future.
	TaskStateRetry

	// TaskStateArchived Indicates that the task is archived and stored for inspection purposes.
	TaskStateArchived

	// TaskStateCompleted Indicates that the task is processed successfully and retained until the retention TTL expires.
	TaskStateCompleted
)

func (s TaskState) String() string {
	switch s {
	case TaskStateActive:
		return "active"
	case TaskStatePending:
		return "pending"
	case TaskStateScheduled:
		return "scheduled"
	case TaskStateRetry:
		return "retry"
	case TaskStateArchived:
		return "archived"
	case TaskStateCompleted:
		return "completed"
	}
	panic("go-kit: unknown task state")
}
