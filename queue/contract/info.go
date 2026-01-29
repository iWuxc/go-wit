package contract

import "time"

// QueueInfo represents a state of a queue at a certain time.
type QueueInfo struct {
	// Name of the queue.
	Name string

	// Latency of the queue, measured by the oldest pending task in the queue.
	Latency time.Duration

	// Total is the total number of tasks in the queue.
	// The value is the sum of Pending, Active, Scheduled, Retry and Archived.
	Total int

	// Number of pending tasks.
	Pending int
	// Number of active tasks.
	Active int
	// Number of scheduled tasks.
	Scheduled int
	// Number of retry tasks.
	Retry int
	// Number of archived tasks.
	Archived int
	// Number of stored completed tasks.
	Completed int

	// Total number of tasks being processed within the given date (counter resets daily).
	// The number includes both succeeded and failed tasks.
	Processed int
	// Total number of tasks failed to be processed within the given date (counter resets daily).
	Failed int

	// Total number of tasks processed (cumulative).
	ProcessedTotal int
	// Total number of tasks failed (cumulative).
	FailedTotal int

	// Paused indicates whether the queue is paused.
	// If true, tasks in the queue will not be processed.
	Paused bool

	// Time when this queue info snapshot was taken.
	Timestamp time.Time
}
