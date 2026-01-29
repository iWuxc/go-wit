package server

import (
	"context"
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/queue/contract"
	"sync"
	"time"
)

// healthchecker is responsible for pinging broker periodically
// and call user provided HeathCheckFunc with the ping result.
type healthchecker struct {
	broker contract.BrokerInterface

	// channel to communicate back to the long-running "healthchecker" goroutine.
	done chan struct{}

	// interval between healthcheck.
	interval time.Duration

	// function to call periodically.
	healthcheckFunc func(error)
}

type healthcheckerParams struct {
	broker          contract.BrokerInterface
	interval        time.Duration
	healthcheckFunc func(error)
}

func newHealthChecker(params healthcheckerParams) *healthchecker {
	return &healthchecker{
		broker:          params.broker,
		done:            make(chan struct{}),
		interval:        params.interval,
		healthcheckFunc: params.healthcheckFunc,
	}
}

func (hc *healthchecker) shutdown() {
	if hc.healthcheckFunc == nil {
		return
	}

	log.Debug("Healthchecker shutting down...")
	// Signal the healthchecker goroutine to stop.
	hc.done <- struct{}{}
}

func (hc *healthchecker) start(wg *sync.WaitGroup) {
	if hc.healthcheckFunc == nil {
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		timer := time.NewTimer(hc.interval)
		for {
			select {
			case <-hc.done:
				log.Debug("Healthchecker done")
				timer.Stop()
				return
			case <-timer.C:
				err := hc.broker.Ping(context.Background())
				hc.healthcheckFunc(err)
				timer.Reset(hc.interval)
			}
		}
	}()
}
