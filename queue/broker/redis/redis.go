package redis

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/queue/contract"
	"github.com/iWuxc/go-wit/redis"
	"github.com/spf13/cast"
	"math"
	"sync"
	"time"
)

const statsTTL = 90 * 24 * time.Hour // 90 days

// LeaseDuration is the duration used to initially create a lease and to extend it thereafter.
const LeaseDuration = 30 * time.Second

var _ contract.BrokerInterface = (*Broker)(nil)

type Broker struct {
	client *redis.Redis
}

// NewRedisBroker .
func NewRedisBroker() *Broker {
	return NewRedisBrokerWithClient(redis.GetRedis())
}

func NewRedisBrokerWithClient(client *redis.Redis) *Broker {
	return &Broker{client: client}
}

func (b *Broker) Ping(ctx context.Context) error {
	return b.client.Ping(ctx)
}

func (b *Broker) Enqueue(ctx context.Context, msg *contract.Task) error {
	encoded, err := contract.EncodeMessage(msg)
	if err != nil {
		return errors.Errorf(fmt.Sprintf("cannot encode message: %v", err))
	}
	if _, err := b.client.SAdd(ctx, contract.AllQueues, msg.Queue); err != nil {
		return &errors.RedisError{Command: "sadd", Err: err}
	}
	keys := []string{
		contract.TaskKey(msg.Queue, msg.ID),
		contract.PendingKey(msg.Queue),
	}
	argv := []interface{}{
		encoded,
		msg.ID,
		time.Now().UnixNano(),
	}
	n, err := b.client.Scripts(ctx, enqueue, keys, argv...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.ErrTaskIdConflict
	}
	return nil
}

func (b *Broker) EnqueueUnique(ctx context.Context, msg *contract.Task, ttl time.Duration) error {
	encoded, err := contract.EncodeMessage(msg)
	if err != nil {
		return errors.Errorf("cannot encode task message: %v", err)
	}
	if _, err := b.client.SAdd(ctx, contract.AllQueues, msg.Queue); err != nil {
		return &errors.RedisError{Command: "sadd", Err: err}
	}
	keys := []string{
		msg.UniqueKey,
		contract.TaskKey(msg.Queue, msg.ID),
		contract.PendingKey(msg.Queue),
	}
	argv := []interface{}{
		msg.ID,
		int(ttl.Seconds()),
		encoded,
		time.Now().UnixNano(),
	}
	n, err := b.client.Scripts(ctx, enqueueUnique, keys, argv...)
	if err != nil {
		return err
	}
	if n == -1 {
		return errors.ErrDuplicateTask
	}
	if n == 0 {
		return errors.ErrTaskIdConflict
	}
	return nil
}

func (b *Broker) Dequeue(queues ...string) (msg *contract.Task, leaseExpirationTime time.Time, err error) {
	for _, qname := range queues {
		keys := []string{
			contract.PendingKey(qname),
			contract.PausedKey(qname),
			contract.ActiveKey(qname),
			contract.LeaseKey(qname),
		}
		leaseExpirationTime = time.Now().Add(LeaseDuration)
		argv := []interface{}{
			leaseExpirationTime.Unix(),
			contract.TaskKeyPrefix(qname),
		}
		res, err := b.client.Scripts(context.Background(), dequeue, keys, argv...)
		if err == redis.Nil {
			continue
		} else if err != nil {
			return nil, time.Time{}, errors.Errorf(fmt.Sprintf("redis eval error: %v", err))
		}
		encoded, err := cast.ToStringE(res)
		if err != nil {
			return nil, time.Time{}, errors.Errorf(fmt.Sprintf("cast error: unexpected return value from Lua script: %v", res))
		}
		if msg, err = contract.DecodeMessage([]byte(encoded)); err != nil {
			return nil, time.Time{}, errors.Errorf(fmt.Sprintf("cannot decode message: %v", err))
		}
		return msg, leaseExpirationTime, nil
	}
	return nil, time.Time{}, errors.ErrNoProcessableTask
}

func (b *Broker) Done(ctx context.Context, msg *contract.Task) error {
	now := time.Now()
	expireAt := now.Add(statsTTL)
	keys := []string{
		contract.ActiveKey(msg.Queue),
		contract.LeaseKey(msg.Queue),
		contract.TaskKey(msg.Queue, msg.ID),
		contract.ProcessedKey(msg.Queue, now),
		contract.ProcessedTotalKey(msg.Queue),
	}
	argv := []interface{}{
		msg.ID,
		expireAt.Unix(),
		math.MaxInt64,
	}
	// Note: We cannot pass empty unique key when running this script in redis-cluster.
	if len(msg.UniqueKey) > 0 {
		keys = append(keys, msg.UniqueKey)
		_, err := b.client.Scripts(ctx, doneUnique, keys, argv...)
		return err
	}
	_, err := b.client.Scripts(ctx, done, keys, argv...)
	return err
}

func (b *Broker) MarkAsComplete(ctx context.Context, msg *contract.Task) error {
	now := time.Now()
	statsExpireAt := now.Add(statsTTL)
	msg.CompletedAt = now.Unix()
	encoded, err := contract.EncodeMessage(msg)
	if err != nil {
		return errors.Errorf(fmt.Sprintf("cannot encode message: %v", err))
	}
	keys := []string{
		contract.ActiveKey(msg.Queue),
		contract.LeaseKey(msg.Queue),
		contract.CompletedKey(msg.Queue),
		contract.TaskKey(msg.Queue, msg.ID),
		contract.ProcessedKey(msg.Queue, now),
		contract.ProcessedTotalKey(msg.Queue),
	}
	argv := []interface{}{
		msg.ID,
		statsExpireAt.Unix(),
		now.Unix() + msg.Retention,
		encoded,
		math.MaxInt64,
	}
	// Note: We cannot pass empty unique key when running this script in redis-cluster.
	if len(msg.UniqueKey) > 0 {
		keys = append(keys, msg.UniqueKey)
		_, err := b.client.Scripts(ctx, markAsCompleteUnique, keys, argv...)
		return err
	}
	_, err = b.client.Scripts(ctx, markAsComplete, keys, argv...)
	return err
}

func (b *Broker) Requeue(ctx context.Context, msg *contract.Task) error {
	keys := []string{
		contract.ActiveKey(msg.Queue),
		contract.LeaseKey(msg.Queue),
		contract.PendingKey(msg.Queue),
		contract.TaskKey(msg.Queue, msg.ID),
	}
	_, err := b.client.Scripts(ctx, requeue, keys, msg.ID)
	return err
}

func (b *Broker) Schedule(ctx context.Context, msg *contract.Task, processAt time.Time) error {
	encoded, err := contract.EncodeMessage(msg)
	if err != nil {
		return errors.Errorf(fmt.Sprintf("cannot encode message: %v", err))
	}
	if _, err := b.client.SAdd(ctx, contract.AllQueues, msg.Queue); err != nil {
		return &errors.RedisError{Command: "sadd", Err: err}
	}
	keys := []string{
		contract.TaskKey(msg.Queue, msg.ID),
		contract.ScheduledKey(msg.Queue),
	}
	argv := []interface{}{
		encoded,
		processAt.Unix(),
		msg.ID,
	}

	n, err := b.client.Scripts(ctx, schedule, keys, argv...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.ErrTaskIdConflict
	}
	return nil
}

func (b *Broker) ScheduleUnique(ctx context.Context, task *contract.Task, processAt time.Time, ttl time.Duration) error {
	encoded, err := contract.EncodeMessage(task)
	if err != nil {
		return errors.Errorf(fmt.Sprintf("cannot encode task message: %v", err))
	}
	if _, err := b.client.SAdd(ctx, contract.AllQueues, task.Queue); err != nil {
		return &errors.RedisError{Command: "sadd", Err: err}
	}
	keys := []string{
		task.UniqueKey,
		contract.TaskKey(task.Queue, task.ID),
		contract.ScheduledKey(task.Queue),
	}
	argv := []interface{}{
		task.ID,
		int(ttl.Seconds()),
		processAt.Unix(),
		encoded,
	}

	n, err := b.client.Scripts(ctx, scheduleUnique, keys, argv...)
	if err != nil {
		return err
	}
	if n == -1 {
		return errors.ErrDuplicateTask
	}
	if n == 0 {
		return errors.ErrTaskIdConflict
	}
	return nil
}

func (b *Broker) Retry(ctx context.Context, task *contract.Task, processAt time.Time, errMsg string, isFailure bool) error {
	now := time.Now()
	modified := *task
	if isFailure {
		modified.Retried++
	}
	modified.ErrorMsg = errMsg
	modified.LastFailedAt = now.Unix()
	encoded, err := contract.EncodeMessage(&modified)
	if err != nil {
		return errors.Errorf(fmt.Sprintf("cannot encode message: %v", err))
	}
	expireAt := now.Add(statsTTL)
	keys := []string{
		contract.TaskKey(task.Queue, task.ID),
		contract.ActiveKey(task.Queue),
		contract.LeaseKey(task.Queue),
		contract.RetryKey(task.Queue),
		contract.ProcessedKey(task.Queue, now),
		contract.FailedKey(task.Queue, now),
		contract.ProcessedTotalKey(task.Queue),
		contract.FailedTotalKey(task.Queue),
	}
	argv := []interface{}{
		task.ID,
		encoded,
		processAt.Unix(),
		expireAt.Unix(),
		isFailure,
		math.MaxInt64,
	}

	_, e := b.client.Scripts(ctx, retry, keys, argv...)
	return e
}

const (
	maxArchiveSize           = 10000 // maximum number of tasks in archive
	archivedExpirationInDays = 90    // number of days before an archived task gets deleted permanently
)

func (b *Broker) Archive(ctx context.Context, task *contract.Task, errMsg string) error {
	now := time.Now()
	modified := *task
	modified.ErrorMsg = errMsg
	modified.LastFailedAt = now.Unix()
	encoded, err := contract.EncodeMessage(&modified)
	if err != nil {
		return errors.Errorf(fmt.Sprintf("cannot encode message: %v", err))
	}
	cutoff := now.AddDate(0, 0, -archivedExpirationInDays)
	expireAt := now.Add(statsTTL)
	keys := []string{
		contract.TaskKey(task.Queue, task.ID),
		contract.ActiveKey(task.Queue),
		contract.LeaseKey(task.Queue),
		contract.ArchivedKey(task.Queue),
		contract.ProcessedKey(task.Queue, now),
		contract.FailedKey(task.Queue, now),
		contract.ProcessedTotalKey(task.Queue),
		contract.FailedTotalKey(task.Queue),
	}
	argv := []interface{}{
		task.ID,
		encoded,
		now.Unix(),
		cutoff.Unix(),
		maxArchiveSize,
		expireAt.Unix(),
		math.MaxInt64,
	}
	_, e := b.client.Scripts(ctx, archive, keys, argv...)
	return e
}

func (b *Broker) ForwardIfReady(queues ...string) error {
	for _, qname := range queues {
		if err := b.forwardAll(qname); err != nil {
			return err
		}
	}
	return nil
}

// forward moves tasks with a score less than the current unix time
// from the src zset to the dst list. It returns the number of tasks moved.
func (b *Broker) forward(src, dst, taskKeyPrefix string) (int, error) {
	now := time.Now()
	res, err := b.client.Scripts(context.Background(), forward, []string{src, dst}, now.Unix(), taskKeyPrefix, now.UnixNano())
	if err != nil {
		return 0, errors.Errorf(fmt.Sprintf("redis eval error: %v", err))
	}
	n, err := cast.ToIntE(res)
	if err != nil {
		return 0, errors.Errorf(fmt.Sprintf("cast error: Lua script returned unexpected value: %v", res))
	}
	return n, nil
}

// forwardAll checks for tasks in scheduled/retry state that are ready to be run, and updates
// their state to "pending".
func (b *Broker) forwardAll(queue string) (err error) {
	sources := []string{contract.ScheduledKey(queue), contract.RetryKey(queue)}
	dst := contract.PendingKey(queue)
	taskKeyPrefix := contract.TaskKeyPrefix(queue)
	for _, src := range sources {
		n := 1
		for n != 0 {
			n, err = b.forward(src, dst, taskKeyPrefix)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Broker) DeleteExpiredCompletedTasks(queue string) error {
	// Note: Do this operation in fix batches to prevent long-running script.
	const batchSize = 100
	for {
		n, err := b.deleteExpiredCompletedTasks(queue, batchSize)
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
	}
}

// deleteExpiredCompletedTasks runs the lua script to delete expired deleted task with the specified
// batch size. It reports the number of tasks deleted.
func (b *Broker) deleteExpiredCompletedTasks(queue string, batchSize int) (int64, error) {
	keys := []string{contract.CompletedKey(queue)}
	argv := []interface{}{
		time.Now().Unix(),
		contract.TaskKeyPrefix(queue),
		batchSize,
	}
	res, err := b.client.Scripts(context.Background(), deleteExpiredCompletedTasks, keys, argv...)
	if err != nil {
		return 0, errors.Errorf(fmt.Sprintf("redis eval error: %v", err))
	}
	n, ok := res.(int64)
	if !ok {
		return 0, errors.Errorf(fmt.Sprintf("unexpected return value from Lua script: %v", res))
	}
	return n, nil
}

func (b *Broker) ListLeaseExpired(cutoff time.Time, queues ...string) ([]*contract.Task, error) {
	var msgs []*contract.Task
	for _, queue := range queues {

		res, err := b.client.Scripts(context.Background(), listLeaseExpired, []string{contract.LeaseKey(queue)}, cutoff.Unix(), contract.TaskKeyPrefix(queue))
		if err != nil {
			return nil, errors.Errorf(fmt.Sprintf("redis eval error: %v", err))
		}

		data, err := cast.ToStringSliceE(res)
		if err != nil {
			return nil, errors.Errorf(fmt.Sprintf("cast error: Lua script returned unexpected value: %v", res))
		}
		for _, s := range data {
			msg, err := contract.DecodeMessage([]byte(s))
			if err != nil {
				return nil, errors.Errorf(fmt.Sprintf("cannot decode message: %v", err))
			}
			msgs = append(msgs, msg)
		}
	}
	return msgs, nil
}

func (b *Broker) ExtendLease(queue string, ids ...string) (time.Time, error) {
	var rw sync.RWMutex
	expireAt := time.Now().Add(LeaseDuration)
	sz := make(map[string]float64)
	for _, id := range ids {
		rw.Lock()
		sz[id] = float64(expireAt.Unix())
		rw.Unlock()
	}

	_, err := b.client.ZAddXX(context.Background(), contract.LeaseKey(queue), b.client.RedisZ(sz)...)
	if err != nil {
		return time.Time{}, err
	}
	return expireAt, nil
}

func (b *Broker) ClearServerState(host string, pid int, serverID string) error {
	ctx := context.Background()
	sKey := contract.ServerInfoKey(host, pid, serverID)
	wKey := contract.WorkersKey(host, pid, serverID)
	if _, err := b.client.ZRem(ctx, contract.AllServers, sKey); err != nil {
		return &errors.RedisError{Command: "zrem", Err: err}
	}
	if _, err := b.client.ZRem(ctx, contract.AllWorkers, wKey); err != nil {
		return &errors.RedisError{Command: "zrem", Err: err}
	}

	_, err := b.client.Scripts(ctx, clearServerState, []string{sKey, wKey})
	return err
}

func (b *Broker) PublishCancelation(id string) error {
	if _, err := b.client.Publish(context.Background(), contract.CancelChannel, id); err != nil {
		return errors.Errorf("redis sub publish error: %v", err)
	}
	return nil
}

func (b *Broker) WriteResult(queue, taskID string, data []byte) (int, error) {
	taskKey := contract.TaskKey(queue, taskID)
	if _, err := b.client.HSet(context.Background(), taskKey, "result", data); err != nil {
		return 0, &errors.RedisError{Command: "hset", Err: err}
	}
	return len(data), nil
}

func (b *Broker) WriteServerState(info *contract.ServerInfo, workers []*contract.WorkerInfo, ttl time.Duration) error {
	ctx := context.Background()
	bytes, err := contract.EncodeServerInfo(info)
	if err != nil {
		return errors.Errorf(fmt.Sprintf("cannot encode server info: %v", err))
	}
	exp := time.Now().Add(ttl).UTC()
	args := []interface{}{ttl.Seconds(), bytes} // args to the lua script
	for _, w := range workers {
		bytes, err := contract.EncodeWorkerInfo(w)
		if err != nil {
			continue // skip bad data
		}
		args = append(args, w.ID, bytes)
	}
	skey := contract.ServerInfoKey(info.Host, info.PID, info.ServerID)
	wkey := contract.WorkersKey(info.Host, info.PID, info.ServerID)

	if _, err := b.client.ZAdd(ctx, contract.AllServers, b.client.RedisZ(map[string]float64{skey: float64(exp.Unix())})...); err != nil {
		return &errors.RedisError{Command: "sadd", Err: err}
	}
	if _, err := b.client.ZAdd(ctx, contract.AllWorkers, b.client.RedisZ(map[string]float64{wkey: float64(exp.Unix())})...); err != nil {
		return &errors.RedisError{Command: "zadd", Err: err}
	}

	_, e := b.client.Scripts(ctx, writeServerState, []string{skey, wkey}, args...)
	return e
}

func (b *Broker) CancelationPubSub(ctx context.Context, retryTimeout time.Duration, done <-chan struct{}, f func(string)) {
	sub := b.client.Subscribe(ctx, contract.CancelChannel)
	_, err := sub.Receive(ctx)
	if err != nil {
		<-time.After(retryTimeout)
		b.CancelationPubSub(ctx, retryTimeout, done, f)
		return
	}

	cancelCh := sub.Channel()
	for {
		select {
		case <-done:
			sub.Close()
			return
		case msg := <-cancelCh:
			f(msg.Payload)
		}
	}
}

func (b *Broker) Close() error {
	return b.client.Close()
}
