package queue

import (
	"fmt"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/queue/contract"
	"github.com/iWuxc/go-wit/utils"
	"strings"
	"time"
)

type OptionType int

const (
	MaxRetryOpt OptionType = iota
	QueueOpt
	TimeoutOpt
	DeadlineOpt
	UniqueOpt
	ProcessAtOpt
	ProcessInOpt
	TaskIDOpt
	RetentionOpt
)

// OptionInterface specifies the task processing behavior.
type OptionInterface interface {
	// String returns a string representation of the option.
	String() string

	// Type describes the type of the option.
	Type() OptionType

	// Value returns a value used to create this option.
	Value() interface{}
}

// Internal option representations.
type (
	retryOption     int
	queueOption     string
	taskIDOption    string
	timeoutOption   time.Duration
	deadlineOption  time.Time
	uniqueOption    time.Duration
	processAtOption time.Time
	processInOption time.Duration
	retentionOption time.Duration
)

// OptMaxRetry returns an option to specify the max number of times
// the task will be retried.
//
// Negative retry count is treated as zero retry.
func OptMaxRetry(n int) OptionInterface {
	if n < 0 {
		n = 0
	}
	return retryOption(n)
}

func (n retryOption) String() string     { return fmt.Sprintf("MaxRetry(%d)", int(n)) }
func (n retryOption) Type() OptionType   { return MaxRetryOpt }
func (n retryOption) Value() interface{} { return int(n) }

// OptQueue returns an option to specify the queue to enqueue the task into.
func OptQueue(qname string) OptionInterface {
	return queueOption(qname)
}

func (qname queueOption) String() string     { return fmt.Sprintf("Queue(%q)", string(qname)) }
func (qname queueOption) Type() OptionType   { return QueueOpt }
func (qname queueOption) Value() interface{} { return string(qname) }

// OptTaskID returns an option to specify the task ID.
func OptTaskID(id string) OptionInterface {
	return taskIDOption(id)
}

func (id taskIDOption) String() string     { return fmt.Sprintf("TaskID(%q)", string(id)) }
func (id taskIDOption) Type() OptionType   { return TaskIDOpt }
func (id taskIDOption) Value() interface{} { return string(id) }

// OptTimeout returns an option to specify how long a task may run.
// If the timeout elapses before the Handler returns, then the task
// will be retried.
//
// Zero duration means no limit.
//
// If there's a conflicting Deadline option, whichever comes earliest
// will be used.
func OptTimeout(d time.Duration) OptionInterface {
	return timeoutOption(d)
}

func (d timeoutOption) String() string     { return fmt.Sprintf("Timeout(%v)", time.Duration(d)) }
func (d timeoutOption) Type() OptionType   { return TimeoutOpt }
func (d timeoutOption) Value() interface{} { return time.Duration(d) }

// OptDeadline returns an option to specify the deadline for the given task.
// If it reaches the deadline before the Handler returns, then the task
// will be retried.
//
// If there's a conflicting Timeout option, whichever comes earliest
// will be used.
func OptDeadline(t time.Time) OptionInterface {
	return deadlineOption(t)
}

func (t deadlineOption) String() string {
	return fmt.Sprintf("Deadline(%v)", time.Time(t).Format(time.UnixDate))
}
func (t deadlineOption) Type() OptionType   { return DeadlineOpt }
func (t deadlineOption) Value() interface{} { return time.Time(t) }

// OptUnique returns an option to enqueue a task only if the given task is unique.
// Task enqueued with this option is guaranteed to be unique within the given ttl.
// Once the task gets processed successfully or once the TTL has expired, another task with the same uniqueness may be enqueued.
// ErrDuplicateTask error is returned when enqueueing a duplicate task.
// TTL duration must be greater than or equal to 1 second.
//
// Uniqueness of a task is based on the following properties:
//     - Task Type
//     - Task Payload
//     - Queue Name
func OptUnique(ttl time.Duration) OptionInterface {
	return uniqueOption(ttl)
}

func (ttl uniqueOption) String() string     { return fmt.Sprintf("Unique(%v)", time.Duration(ttl)) }
func (ttl uniqueOption) Type() OptionType   { return UniqueOpt }
func (ttl uniqueOption) Value() interface{} { return time.Duration(ttl) }

// OptProcessAt returns an option to specify when to process the given task.
//
// If there's a conflicting ProcessIn option, the last option passed to Enqueue overrides the others.
func OptProcessAt(t time.Time) OptionInterface {
	return processAtOption(t)
}

func (t processAtOption) String() string {
	return fmt.Sprintf("ProcessAt(%v)", time.Time(t).Format(time.UnixDate))
}
func (t processAtOption) Type() OptionType   { return ProcessAtOpt }
func (t processAtOption) Value() interface{} { return time.Time(t) }

// OptProcessIn returns an option to specify when to process the given task relative to the current time.
//
// If there's a conflicting ProcessAt option, the last option passed to Enqueue overrides the others.
func OptProcessIn(d time.Duration) OptionInterface {
	return processInOption(d)
}

func (d processInOption) String() string     { return fmt.Sprintf("ProcessIn(%v)", time.Duration(d)) }
func (d processInOption) Type() OptionType   { return ProcessInOpt }
func (d processInOption) Value() interface{} { return time.Duration(d) }

// OptRetention returns an option to specify the duration of retention period for the task.
// If this option is provided, the task will be stored as a completed task after successful processing.
// A completed task will be deleted after the specified duration elapses.
func OptRetention(d time.Duration) OptionInterface {
	return retentionOption(d)
}

func (ttl retentionOption) String() string     { return fmt.Sprintf("Retention(%v)", time.Duration(ttl)) }
func (ttl retentionOption) Type() OptionType   { return RetentionOpt }
func (ttl retentionOption) Value() interface{} { return time.Duration(ttl) }

type Option struct {
	Retry     int
	Queue     string
	TaskID    string
	Timeout   time.Duration
	Deadline  time.Time
	UniqueTTL time.Duration
	ProcessAt time.Time
	Retention time.Duration
}

// ComposeOptions merges user provided options into the default options
// and returns the composed option.
// It also validates the user provided options and returns an error if any of
// the user provided options fail the validations.
func ComposeOptions(opts ...OptionInterface) (Option, error) {
	res := Option{
		Retry:     contract.DefaultMaxRetry,
		Queue:     contract.DefaultQueueName,
		TaskID:    utils.GetUUID(),
		Timeout:   0,
		Deadline:  time.Time{},
		ProcessAt: time.Now(),
	}
	for _, opt := range opts {
		switch opt := opt.(type) {
		case retryOption:
			res.Retry = int(opt)
		case queueOption:
			qname := string(opt)
			if err := contract.ValidateQueueName(qname); err != nil {
				return Option{}, err
			}
			res.Queue = qname
		case taskIDOption:
			id := string(opt)
			if err := validateTaskID(id); err != nil {
				return Option{}, err
			}
			res.TaskID = id
		case timeoutOption:
			res.Timeout = time.Duration(opt)
		case deadlineOption:
			res.Deadline = time.Time(opt)
		case uniqueOption:
			ttl := time.Duration(opt)
			if ttl < 1*time.Second {
				return Option{}, errors.New("Unique TTL cannot be less than 1s")
			}
			res.UniqueTTL = ttl
		case processAtOption:
			res.ProcessAt = time.Time(opt)
		case processInOption:
			res.ProcessAt = time.Now().Add(time.Duration(opt))
		case retentionOption:
			res.Retention = time.Duration(opt)
		default:
			// ignore unexpected option
		}
	}
	return res, nil
}

// validates user provided task ID string.
func validateTaskID(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("task ID cannot be empty")
	}
	return nil
}
