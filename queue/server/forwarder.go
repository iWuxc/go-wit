package server

import (
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/queue/contract"
	"sync"
	"time"
)

// A forwarder is responsible for moving scheduled and retry tasks to pending state
// so that the tasks get processed by the workers.
type forwarder struct {
	broker contract.BrokerInterface

	// channel to communicate back to the long-running "forwarder" goroutine.
	done chan struct{}

	// list of queue names to check and enqueue.
	queues []string

	// poll interval on average
	avgInterval time.Duration
}

type forwarderParams struct {
	broker   contract.BrokerInterface
	queues   []string
	interval time.Duration
}

func newForwarder(params forwarderParams) *forwarder {
	return &forwarder{
		broker:      params.broker,
		done:        make(chan struct{}),
		queues:      params.queues,
		avgInterval: params.interval,
	}
}

func (f *forwarder) shutdown() {
	log.Debug("Forwarder shutting down...")
	// Signal the forwarder goroutine to stop polling.
	f.done <- struct{}{}
}

// start the "forwarder" goroutine.
func (f *forwarder) start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-f.done:
				log.Debug("Forwarder done")
				return
			case <-time.After(f.avgInterval):
				f.exec()
			}
		}
	}()
}

func (f *forwarder) exec() {
	if err := f.broker.ForwardIfReady(f.queues...); err != nil {
		log.Errorf("Failed to forward scheduled tasks: %v", err)
	}
}
