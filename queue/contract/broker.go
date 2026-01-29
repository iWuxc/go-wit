package contract

import (
	"context"
	"time"
)

type BrokerInterface interface {
	Ping(ctx context.Context) error
	Enqueue(ctx context.Context, msg *Task) error
	EnqueueUnique(ctx context.Context, msg *Task, ttl time.Duration) error
	Dequeue(queue ...string) (*Task, time.Time, error)
	Done(ctx context.Context, msg *Task) error
	MarkAsComplete(ctx context.Context, msg *Task) error
	Requeue(ctx context.Context, msg *Task) error
	Schedule(ctx context.Context, msg *Task, processAt time.Time) error
	ScheduleUnique(ctx context.Context, msg *Task, processAt time.Time, ttl time.Duration) error
	Retry(ctx context.Context, msg *Task, processAt time.Time, errMsg string, isFailure bool) error
	Archive(ctx context.Context, msg *Task, errMsg string) error
	ForwardIfReady(queue ...string) error
	DeleteExpiredCompletedTasks(queue string) error
	ListLeaseExpired(cutoff time.Time, queue ...string) ([]*Task, error)
	ExtendLease(queue string, ids ...string) (time.Time, error)
	WriteServerState(info *ServerInfo, workers []*WorkerInfo, ttl time.Duration) error
	ClearServerState(host string, pid int, serverID string) error
	PublishCancelation(id string) error
	WriteResult(queue, id string, data []byte) (int, error)
	CancelationPubSub(ctx context.Context, retryTimeout time.Duration, done <-chan struct{}, f func(string))
	Close() error
}
