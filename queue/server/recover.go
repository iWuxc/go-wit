package server

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/queue"
	"github.com/iWuxc/go-wit/queue/contract"
	"sync"
	"time"
)

type recoverer struct {
	broker         contract.BrokerInterface
	retryDelayFunc RetryDelayFunc
	isFailureFunc  func(error) bool

	// channel to communicate back to the long-running "recoverer" goroutine.
	done chan struct{}

	// list of queues to check for deadline.
	queues []string

	// poll interval.
	interval time.Duration
}

type recovererParams struct {
	broker         contract.BrokerInterface
	queues         []string
	interval       time.Duration
	retryDelayFunc RetryDelayFunc
	isFailureFunc  func(error) bool
}

func newRecoverer(params recovererParams) *recoverer {
	return &recoverer{
		broker:         params.broker,
		done:           make(chan struct{}),
		queues:         params.queues,
		interval:       params.interval,
		retryDelayFunc: params.retryDelayFunc,
		isFailureFunc:  params.isFailureFunc,
	}
}

func (r *recoverer) shutdown() {
	log.Debug("Recoverer shutting down...")
	// Signal the recoverer goroutine to stop polling.
	r.done <- struct{}{}
}

func (r *recoverer) start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.recover()
		timer := time.NewTimer(r.interval)
		for {
			select {
			case <-r.done:
				log.Debug("Recoverer done")
				timer.Stop()
				return
			case <-timer.C:
				r.recover()
				timer.Reset(r.interval)
			}
		}
	}()
}

// ErrLeaseExpired error indicates that the task failed because the worker working on the task
// could not extend its lease due to missing heartbeats. The worker may have crashed or got cutoff from the network.
var ErrLeaseExpired = errors.New("go-kit: task lease expired")

func (r *recoverer) recover() {
	// Get all tasks which have expired 30 seconds ago or earlier to accommodate certain amount of clock skew.
	cutoff := time.Now().Add(-30 * time.Second)
	msgs, err := r.broker.ListLeaseExpired(cutoff, r.queues...)
	if err != nil {
		log.Warn("recoverer: could not list lease expired tasks")
		return
	}
	for _, msg := range msgs {
		if msg.Retried >= msg.Retry {
			r.archive(msg, ErrLeaseExpired)
		} else {
			r.retry(msg, ErrLeaseExpired)
		}
	}
}

func (r *recoverer) retry(msg *contract.Task, err error) {
	delay := r.retryDelayFunc(msg.Retried, err, queue.NewTask(msg.Type, msg.Payload))
	retryAt := time.Now().Add(delay)
	if err := r.broker.Retry(context.Background(), msg, retryAt, err.Error(), r.isFailureFunc(err)); err != nil {
		log.Warnf("recoverer: could not retry lease expired task: %v", err)
	}
}

func (r *recoverer) archive(msg *contract.Task, err error) {
	if err := r.broker.Archive(context.Background(), msg, err.Error()); err != nil {
		log.Warnf("recoverer: could not move task to archive: %v", err)
	}
}
