package contract

import (
	"fmt"
	pb "github.com/iWuxc/go-wit/queue/proto"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"time"
)

// TaskState denotes the state of a task.
type TaskState int

const (
	TaskStateActive TaskState = iota + 1
	TaskStatePending
	TaskStateScheduled
	TaskStateRetry
	TaskStateArchived
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
	panic(fmt.Sprintf("internal error: unknown task state %d", s))
}

func TaskStateFromString(s string) (TaskState, error) {
	switch s {
	case "active":
		return TaskStateActive, nil
	case "pending":
		return TaskStatePending, nil
	case "scheduled":
		return TaskStateScheduled, nil
	case "retry":
		return TaskStateRetry, nil
	case "archived":
		return TaskStateArchived, nil
	case "completed":
		return TaskStateCompleted, nil
	}
	return 0, errors.Errorf("not supported task state: %s", s)
}

// TaskInfo describes a task message and its metadata.
type TaskInfo struct {
	Message       *Task
	State         TaskState
	NextProcessAt time.Time
	Result        []byte
}

type Task struct {
	// Type indicates the kind of the task to be performed.
	Type string

	// Payload holds data needed to process the task.
	Payload []byte

	// ID is a unique identifier for each task.
	ID string

	// Queue is a name this message should be enqueued to.
	Queue string

	// Retry is the max number of retry for this task.
	Retry int

	// Retried is the number of times we've retried this task so far.
	Retried int

	// ErrorMsg holds the error message from the last failure.
	ErrorMsg string

	// Time of last failure in Unix time,
	// the number of seconds elapsed since January 1, 1970, UTC.
	//
	// Use zero to indicate no last failure
	LastFailedAt int64

	// Timeout specifies timeout in seconds.
	// If task processing doesn't complete within the timeout, the task will be retried
	// if retry count is remaining. Otherwise, it will be moved to the archive.
	//
	// Use zero to indicate no timeout.
	Timeout int64

	// Deadline specifies the deadline for the task in Unix time,
	// the number of seconds elapsed since January 1, 1970, UTC.
	// If task processing doesn't complete before the deadline, the task will be retried
	// if retry count is remaining. Otherwise, it will be moved to the archive.
	//
	// Use zero to indicate no deadline.
	Deadline int64

	// UniqueKey holds the redis key used for uniqueness lock for this task.
	//
	// Empty string indicates that no uniqueness lock was used.
	UniqueKey string

	// Retention specifies the number of seconds the task should be retained after completion.
	Retention int64

	// CompletedAt is the time the task was processed successfully in Unix time,
	// the number of seconds elapsed since January 1, 1970, UTC.
	//
	// Use zero to indicate no value.
	CompletedAt int64
}

// EncodeMessage marshals the given task message and returns an encoded bytes.
func EncodeMessage(msg *Task) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("cannot encode nil message")
	}
	return proto.Marshal(&pb.Task{
		Type:         msg.Type,
		Payload:      msg.Payload,
		Id:           msg.ID,
		Queue:        msg.Queue,
		Retry:        int32(msg.Retry),
		Retried:      int32(msg.Retried),
		ErrorMsg:     msg.ErrorMsg,
		LastFailedAt: msg.LastFailedAt,
		Timeout:      msg.Timeout,
		Deadline:     msg.Deadline,
		UniqueKey:    msg.UniqueKey,
		Retention:    msg.Retention,
		CompletedAt:  msg.CompletedAt,
	})
}

// DecodeMessage unmarshal the given bytes and returns a decoded task message.
func DecodeMessage(data []byte) (*Task, error) {
	var pbmsg pb.Task
	if err := proto.Unmarshal(data, &pbmsg); err != nil {
		return nil, err
	}
	return &Task{
		Type:         pbmsg.GetType(),
		Payload:      pbmsg.GetPayload(),
		ID:           pbmsg.GetId(),
		Queue:        pbmsg.GetQueue(),
		Retry:        int(pbmsg.GetRetry()),
		Retried:      int(pbmsg.GetRetried()),
		ErrorMsg:     pbmsg.GetErrorMsg(),
		LastFailedAt: pbmsg.GetLastFailedAt(),
		Timeout:      pbmsg.GetTimeout(),
		Deadline:     pbmsg.GetDeadline(),
		UniqueKey:    pbmsg.GetUniqueKey(),
		Retention:    pbmsg.GetRetention(),
		CompletedAt:  pbmsg.GetCompletedAt(),
	}, nil
}
