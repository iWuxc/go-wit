package contract

import (
	"fmt"
	"github.com/iWuxc/go-wit/utils"
	"strings"
	"time"
)

const (
	// DefaultQueueName is the queue name used if none are specified by user.
	DefaultQueueName = "default"

	// DefaultMaxRetry Default max retry count used if nothing is specified.
	DefaultMaxRetry = 5

	// DefaultTimeout Default timeout used if both timeout and deadline are not specified.
	DefaultTimeout = 30 * time.Minute
)

// DefaultQueue is the redis key for the default queue.
var DefaultQueue = PendingKey(DefaultQueueName)

// Global Redis keys.
const (
	AllServers    = "go-kit:servers"    // ZSET
	AllWorkers    = "go-kit:workers"    // ZSET
	AllSchedulers = "go-kit:schedulers" // ZSET
	AllQueues     = "go-kit:queues"     // SET
	CancelChannel = "go-kit:cancel"     // PubSub channel
)

// ValidateQueueName validates a given queue to be used as a queue name.
// Returns nil if valid, otherwise returns non-nil error.
func ValidateQueueName(queue string) error {
	if len(strings.TrimSpace(queue)) == 0 {
		return fmt.Errorf("queue name must contain one or more characters")
	}
	return nil
}

// QueueKeyPrefix returns a prefix for all keys in the given queue.
func QueueKeyPrefix(queue string) string {
	return fmt.Sprintf("go-kit:{%s}:", queue)
}

// PendingKey returns a redis key for the given queue name.
func PendingKey(queue string) string {
	return fmt.Sprintf("%spending", QueueKeyPrefix(queue))
}

// TaskKeyPrefix returns a prefix for task key.
func TaskKeyPrefix(queue string) string {
	return fmt.Sprintf("%st:", QueueKeyPrefix(queue))
}

// TaskKey returns a redis key for the given task message.
func TaskKey(queue, id string) string {
	return fmt.Sprintf("%s%s", TaskKeyPrefix(queue), id)
}

// ActiveKey returns a redis key for the active tasks.
func ActiveKey(queue string) string {
	return fmt.Sprintf("%sactive", QueueKeyPrefix(queue))
}

// ScheduledKey returns a redis key for the scheduled tasks.
func ScheduledKey(queue string) string {
	return fmt.Sprintf("%sscheduled", QueueKeyPrefix(queue))
}

// RetryKey returns a redis key for the retry tasks.
func RetryKey(queue string) string {
	return fmt.Sprintf("%sretry", QueueKeyPrefix(queue))
}

// ArchivedKey returns a redis key for the archived tasks.
func ArchivedKey(queue string) string {
	return fmt.Sprintf("%sarchived", QueueKeyPrefix(queue))
}

// LeaseKey returns a redis key for the lease.
func LeaseKey(queue string) string {
	return fmt.Sprintf("%slease", QueueKeyPrefix(queue))
}

func CompletedKey(queue string) string {
	return fmt.Sprintf("%scompleted", QueueKeyPrefix(queue))
}

// PausedKey returns a redis key to indicate that the given queue is paused.
func PausedKey(queue string) string {
	return fmt.Sprintf("%spaused", QueueKeyPrefix(queue))
}

// ProcessedTotalKey returns a redis key for total processed count for the given queue.
func ProcessedTotalKey(queue string) string {
	return fmt.Sprintf("%sprocessed", QueueKeyPrefix(queue))
}

// FailedTotalKey returns a redis key for total failure count for the given queue.
func FailedTotalKey(queue string) string {
	return fmt.Sprintf("%sfailed", QueueKeyPrefix(queue))
}

// ProcessedKey returns a redis key for processed count for the given day for the queue.
func ProcessedKey(queue string, t time.Time) string {
	return fmt.Sprintf("%sprocessed:%s", QueueKeyPrefix(queue), t.UTC().Format("2006-01-02"))
}

// FailedKey returns a redis key for failure count for the given day for the queue.
func FailedKey(queue string, t time.Time) string {
	return fmt.Sprintf("%sfailed:%s", QueueKeyPrefix(queue), t.UTC().Format("2006-01-02"))
}

// ServerInfoKey returns a redis key for process info.
func ServerInfoKey(hostname string, pid int, serverID string) string {
	return fmt.Sprintf("go-kit:servers:{%s:%d:%s}", hostname, pid, serverID)
}

// WorkersKey returns a redis key for the workers given hostname, pid, and server ID.
func WorkersKey(hostname string, pid int, serverID string) string {
	return fmt.Sprintf("go-kit:workers:{%s:%d:%s}", hostname, pid, serverID)
}

// SchedulerEntriesKey returns a redis key for the scheduler entries given scheduler ID.
func SchedulerEntriesKey(schedulerID string) string {
	return fmt.Sprintf("go-kit:schedulers:{%s}", schedulerID)
}

// SchedulerHistoryKey returns a redis key for the scheduler's history for the given entry.
func SchedulerHistoryKey(entryID string) string {
	return fmt.Sprintf("go-kit:scheduler_history:%s", entryID)
}

// UniqueKey returns a redis key with the given type, payload, and queue name.
func UniqueKey(queue, tasktype string, payload []byte) string {
	if payload == nil {
		return fmt.Sprintf("%sunique:%s:", QueueKeyPrefix(queue), tasktype)
	}

	return fmt.Sprintf("%sunique:%s:%s", QueueKeyPrefix(queue), tasktype, utils.HashMd5ForByte(payload))
}
