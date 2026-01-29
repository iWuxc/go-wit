package server

import (
	"context"
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/queue/contract"
	"sync"
	"time"
)

type subscriber struct {
	broker contract.BrokerInterface

	// channel to communicate back to the long-running "subscriber" goroutine.
	done chan struct{}

	// cancellations hold cancel functions for all active tasks.
	cancellations *contract.Cancellations

	// time to wait before retrying to connect to redis.
	retryTimeout time.Duration
}

type subscriberParams struct {
	broker        contract.BrokerInterface
	cancellations *contract.Cancellations
}

func newSubscriber(params subscriberParams) *subscriber {
	return &subscriber{
		broker:        params.broker,
		done:          make(chan struct{}),
		cancellations: params.cancellations,
		retryTimeout:  5 * time.Second,
	}
}

func (s *subscriber) shutdown() {
	log.Debug("Subscriber shutting down...")
	// Signal the subscriber goroutine to stop.
	s.done <- struct{}{}
}

func (s *subscriber) start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.broker.CancelationPubSub(context.Background(), s.retryTimeout, s.done,
			func(str string) {
				cancel, ok := s.cancellations.Get(str)
				if ok {
					cancel()
				}
			})
	}()
}
