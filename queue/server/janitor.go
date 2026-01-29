package server

import (
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/queue/contract"
	"sync"
	"time"
)

// A janitor is responsible for deleting expired completed tasks from the specified
// queues. It periodically checks for any expired tasks in the completed set, and
// deletes them.
type janitor struct {
	broker contract.BrokerInterface

	// channel to communicate back to the long-running "janitor" goroutine.
	done chan struct{}

	// list of queue names to check.
	queues []string

	// average interval between checks.
	avgInterval time.Duration
}

type janitorParams struct {
	broker   contract.BrokerInterface
	queues   []string
	interval time.Duration
}

func newJanitor(params janitorParams) *janitor {
	return &janitor{
		broker:      params.broker,
		done:        make(chan struct{}),
		queues:      params.queues,
		avgInterval: params.interval,
	}
}

func (j *janitor) shutdown() {
	log.Debug("Janitor shutting down...")
	// Signal the janitor goroutine to stop.
	j.done <- struct{}{}
}

// start the "janitor" goroutine.
func (j *janitor) start(wg *sync.WaitGroup) {
	wg.Add(1)
	timer := time.NewTimer(j.avgInterval) // randomize this interval with margin of 1s
	go func() {
		defer wg.Done()
		for {
			select {
			case <-j.done:
				log.Debug("Janitor done")
				return
			case <-timer.C:
				j.exec()
				timer.Reset(j.avgInterval)
			}
		}
	}()
}

func (j *janitor) exec() {
	for _, qname := range j.queues {
		if err := j.broker.DeleteExpiredCompletedTasks(qname); err != nil {
			log.Errorf("Failed to delete expired completed tasks from queue %q: %v",
				qname, err)
		}
	}
}
