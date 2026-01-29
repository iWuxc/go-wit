package client

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/queue"
	"github.com/iWuxc/go-wit/queue/broker"
	"github.com/iWuxc/go-wit/queue/contract"
	"strings"
	"sync"
	"time"
)

// Value zero indicates no timeout and no deadline.
var (
	noTimeout  time.Duration = 0
	noDeadline               = time.Unix(0, 0)
	_client    *Client
	once       sync.Once
)

type Client struct {
	broker contract.BrokerInterface
}

func NewClient() *Client {
	b := broker.Broker()
	return &Client{b}
}

func NewClientWithBroker(b contract.BrokerInterface) *Client {
	return &Client{b}
}

// Enqueue enqueues the given task to a queue.
//
// Enqueue returns TaskInfo and nil error if the task is enqueued successfully, otherwise returns a non-nil error.
//
// The argument opts specifies the behavior of task processing.
// If there are conflicting Option values the last one overrides others.
// Any options provided to NewTask can be overridden by options passed to Enqueue.
// By default, max retry is set to 25 and timeout is set to 30 minutes.
//
// If no ProcessAt or ProcessIn options are provided, the task will be pending immediately.
//
// Enqueue uses context.Background internally; to specify the context, use EnqueueContext.
func (c *Client) Enqueue(task *queue.Task, opts ...queue.OptionInterface) (*queue.TaskInfo, error) {
	return c.EnqueueContext(context.Background(), task, opts...)
}

func Enqueue(task *queue.Task, opts ...queue.OptionInterface) (*queue.TaskInfo, error) {
	initClient()
	return _client.Enqueue(task, opts...)
}

// EnqueueContext enqueues the given task to a queue.
//
// EnqueueContext returns TaskInfo and nil error if the task is enqueued successfully, otherwise returns a non-nil error.
//
// The argument opts specifies the behavior of task processing.
// If there are conflicting Option values the last one overrides others.
// Any options provided to NewTask can be overridden by options passed to Enqueue.
// By default, max retry is set to 25 and timeout is set to 30 minutes.
//
// If no ProcessAt or ProcessIn options are provided, the task will be pending immediately.
//
// The first argument context applies to the enqueue operation. To specify task timeout and deadline, use Timeout and Deadline option instead.
func (c *Client) EnqueueContext(ctx context.Context, task *queue.Task, opts ...queue.OptionInterface) (*queue.TaskInfo, error) {
	if strings.TrimSpace(task.Type()) == "" {
		return nil, fmt.Errorf("task typename cannot be empty")
	}
	// merge task options with the options provided at enqueue time.
	opts = append(task.Opts, opts...)
	opt, err := queue.ComposeOptions(opts...)
	if err != nil {
		return nil, err
	}
	deadline := noDeadline
	if !opt.Deadline.IsZero() {
		deadline = opt.Deadline
	}
	timeout := noTimeout
	if opt.Timeout != 0 {
		timeout = opt.Timeout
	}
	if deadline.Equal(noDeadline) && timeout == noTimeout {
		// If neither deadline nor timeout are set, use default timeout.
		timeout = contract.DefaultTimeout
	}
	var uniqueKey string
	if opt.UniqueTTL > 0 {
		uniqueKey = contract.UniqueKey(opt.Queue, task.Type(), task.Payload())
	}
	msg := &contract.Task{
		ID:        opt.TaskID,
		Type:      task.Type(),
		Payload:   task.Payload(),
		Queue:     opt.Queue,
		Retry:     opt.Retry,
		Deadline:  deadline.Unix(),
		Timeout:   int64(timeout.Seconds()),
		UniqueKey: uniqueKey,
		Retention: int64(opt.Retention.Seconds()),
	}
	now := time.Now()
	var state contract.TaskState
	if opt.ProcessAt.Before(now) || opt.ProcessAt.Equal(now) {
		opt.ProcessAt = now
		err = c.enqueue(ctx, msg, opt.UniqueTTL)
		state = contract.TaskStatePending
	} else {
		err = c.schedule(ctx, msg, opt.ProcessAt, opt.UniqueTTL)
		state = contract.TaskStateScheduled
	}
	switch {
	case errors.Is(err, errors.ErrDuplicateTask):
		return nil, fmt.Errorf("%w", ErrDuplicateTask)
	case errors.Is(err, errors.ErrTaskIdConflict):
		return nil, fmt.Errorf("%w", ErrTaskIDConflict)
	case err != nil:
		return nil, err
	}
	return queue.NewTaskInfo(msg, state, opt.ProcessAt, nil), nil
}

func EnqueueContext(ctx context.Context, task *queue.Task, opts ...queue.OptionInterface) (*queue.TaskInfo, error) {
	initClient()
	return _client.EnqueueContext(ctx, task, opts...)
}

func (c *Client) enqueue(ctx context.Context, msg *contract.Task, uniqueTTL time.Duration) error {
	if uniqueTTL > 0 {
		return c.broker.EnqueueUnique(ctx, msg, uniqueTTL)
	}
	return c.broker.Enqueue(ctx, msg)
}

func (c *Client) schedule(ctx context.Context, msg *contract.Task, t time.Time, uniqueTTL time.Duration) error {
	if uniqueTTL > 0 {
		ttl := t.Add(uniqueTTL).Sub(time.Now())
		return c.broker.ScheduleUnique(ctx, msg, t, ttl)
	}
	return c.broker.Schedule(ctx, msg, t)
}

func Close() error {
	return _client.Close()
}

// Close closes the connection with redis.
func (c *Client) Close() error {
	return c.broker.Close()
}

func initClient() {
	once.Do(func() {
		_client = NewClient()
	})
}
