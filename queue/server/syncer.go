package server

import (
	"github.com/iWuxc/go-wit/log"
	"sync"
	"time"
)

// syncer is responsible for queuing up failed requests to redis and retry
// those requests to sync state between the background process and redis.
type syncer struct {
	requestsCh <-chan *syncRequest

	// channel to communicate back to the long-running "syncer" goroutine.
	done chan struct{}

	// interval between sync operations.
	interval time.Duration
}

type syncRequest struct {
	fn       func() error // sync operation
	errMsg   string       // error message
	deadline time.Time    // request should be dropped if deadline has been exceeded
}

type syncerParams struct {
	requestsCh <-chan *syncRequest
	interval   time.Duration
}

func newSyncer(params syncerParams) *syncer {
	return &syncer{
		requestsCh: params.requestsCh,
		done:       make(chan struct{}),
		interval:   params.interval,
	}
}

func (s *syncer) shutdown() {
	log.Debug("Syncer shutting down...")
	// Signal the syncer goroutine to stop.
	s.done <- struct{}{}
}

func (s *syncer) start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		var requests []*syncRequest
		for {
			select {
			case <-s.done:
				// Try sync one last time before shutting down.
				for _, req := range requests {
					if err := req.fn(); err != nil {
						log.Error(req.errMsg)
					}
				}
				log.Debug("Syncer done")
				return
			case req := <-s.requestsCh:
				requests = append(requests, req)
			case <-time.After(s.interval):
				var temp []*syncRequest
				for _, req := range requests {
					if req.deadline.Before(time.Now()) {
						continue // drop stale request
					}
					if err := req.fn(); err != nil {
						temp = append(temp, req)
					}
				}
				requests = temp
			}
		}
	}()
}
