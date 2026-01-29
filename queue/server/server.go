package server

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/queue"
	"github.com/iWuxc/go-wit/queue/broker"
	"github.com/iWuxc/go-wit/queue/contract"
	"runtime"
	"sync"
	"time"
)

// Server is responsible for task processing and task lifecycle management.
//
// Server pulls tasks off queues and processes them.
// If the processing of a task is unsuccessful, server will schedule it for a retry.
//
// A task will be retried until either the task gets processed successfully
// or until it reaches its max retry count.
//
// If a task exhausts its retries, it will be moved to the archive and
// will be kept in the archive set.
// Note that the archive size is finite and once it reaches its max size,
// the oldest tasks in the archive will be deleted.
type Server struct {
	broker contract.BrokerInterface

	state *serverState

	// wait group to wait for all goroutines to finish.
	wg            sync.WaitGroup
	forwarder     *forwarder
	processor     *processor
	syncer        *syncer
	heartbeat     *heartbeat
	subscriber    *subscriber
	recoverer     *recoverer
	healthchecker *healthchecker
	janitor       *janitor
}

type RedisConnOpt interface {
	// MakeRedisClient returns a new redis client instance.
	// Return value is intentionally opaque to hide the implementation detail of redis client.
	MakeRedisClient() interface{}
}

// NewServer returns a new Server given a redis connection option
// and server configuration.
func NewServerWithBroker(brk contract.BrokerInterface, cfg Config) *Server {

	b := broker.Broker()
	if brk != nil {
		b = brk
	}

	baseCtxFn := cfg.BaseContext
	if baseCtxFn == nil {
		baseCtxFn = context.Background
	}
	n := cfg.Concurrency
	if n < 1 {
		n = runtime.NumCPU()
	}
	delayFunc := cfg.RetryDelayFunc
	if delayFunc == nil {
		delayFunc = DefaultRetryDelayFunc
	}
	isFailureFunc := cfg.IsFailure
	if isFailureFunc == nil {
		isFailureFunc = defaultIsFailureFunc
	}
	queues := make(map[string]int)
	for qname, p := range cfg.Queues {
		if err := contract.ValidateQueueName(qname); err != nil {
			continue // ignore invalid queue names
		}
		if p > 0 {
			queues[qname] = p
		}
	}
	if len(queues) == 0 {
		queues = defaultQueueConfig
	}
	var qnames []string
	for q := range queues {
		qnames = append(qnames, q)
	}
	shutdownTimeout := cfg.ShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = defaultShutdownTimeout
	}
	healthcheckInterval := cfg.HealthCheckInterval
	if healthcheckInterval == 0 {
		healthcheckInterval = defaultHealthCheckInterval
	}

	starting := make(chan *workerInfo)
	finished := make(chan *contract.Task)
	syncCh := make(chan *syncRequest)
	srvState := &serverState{value: srvStateNew}
	cancels := contract.NewCancellations()

	syncer := newSyncer(syncerParams{
		requestsCh: syncCh,
		interval:   5 * time.Second,
	})
	heartbeat := newHeartbeat(heartbeatParams{
		broker:         b,
		interval:       5 * time.Second,
		concurrency:    n,
		queues:         queues,
		strictPriority: cfg.StrictPriority,
		state:          srvState,
		starting:       starting,
		finished:       finished,
	})
	delayedTaskCheckInterval := cfg.DelayedTaskCheckInterval
	if delayedTaskCheckInterval == 0 {
		delayedTaskCheckInterval = defaultDelayedTaskCheckInterval
	}
	forwarder := newForwarder(forwarderParams{
		broker:   b,
		queues:   qnames,
		interval: delayedTaskCheckInterval,
	})
	subscriber := newSubscriber(subscriberParams{
		broker:        b,
		cancellations: cancels,
	})
	processor := newProcessor(processorParams{
		broker:          b,
		retryDelayFunc:  delayFunc,
		baseCtxFn:       baseCtxFn,
		isFailureFunc:   isFailureFunc,
		syncCh:          syncCh,
		cancellations:   cancels,
		concurrency:     n,
		queues:          queues,
		strictPriority:  cfg.StrictPriority,
		errHandler:      cfg.ErrorHandler,
		shutdownTimeout: shutdownTimeout,
		starting:        starting,
		finished:        finished,
	})
	recoverer := newRecoverer(recovererParams{
		broker:         b,
		retryDelayFunc: delayFunc,
		isFailureFunc:  isFailureFunc,
		queues:         qnames,
		interval:       1 * time.Minute,
	})
	healthchecker := newHealthChecker(healthcheckerParams{
		broker:          b,
		interval:        healthcheckInterval,
		healthcheckFunc: cfg.HealthCheckFunc,
	})
	janitor := newJanitor(janitorParams{
		broker:   b,
		queues:   qnames,
		interval: 8 * time.Second,
	})
	return &Server{
		broker:        b,
		state:         srvState,
		forwarder:     forwarder,
		processor:     processor,
		syncer:        syncer,
		heartbeat:     heartbeat,
		subscriber:    subscriber,
		recoverer:     recoverer,
		healthchecker: healthchecker,
		janitor:       janitor,
	}
}
func NewServer(cfg Config) *Server {
	return NewServerWithBroker(nil, cfg)
}

// Run starts the task processing and blocks until
// an os signal to exit the program is received. Once it receives
// a signal, it gracefully shuts down all active workers and other
// goroutines to process the tasks.
//
// Run returns any error encountered at server startup time.
// If the server has already been shutdown, ErrServerClosed is returned.
func (srv *Server) Run(handler queue.Handler) error {
	if err := srv.Start(handler); err != nil {
		return err
	}
	srv.waitForSignals()
	srv.Shutdown()
	return nil
}

// Start starts the worker server. Once the server has started,
// it pulls tasks off queues and starts a worker goroutine for each task
// and then call Handler to process it.
// Tasks are processed concurrently by the workers up to the number of
// concurrency specified in Config.Concurrency.
//
// Start returns any error encountered at server startup time.
// If the server has already been shutdown, ErrServerClosed is returned.
func (srv *Server) Start(handler queue.Handler) error {
	if handler == nil {
		return fmt.Errorf("go-kit: server cannot run with nil handler")
	}
	srv.processor.handler = handler

	if err := srv.start(); err != nil {
		return err
	}

	srv.heartbeat.start(&srv.wg)
	srv.healthchecker.start(&srv.wg)
	srv.subscriber.start(&srv.wg)
	srv.syncer.start(&srv.wg)
	srv.recoverer.start(&srv.wg)
	srv.forwarder.start(&srv.wg)
	srv.processor.start(&srv.wg)
	srv.janitor.start(&srv.wg)
	return nil
}

// Checks server state and returns an error if pre-condition is not met.
// Otherwise, it sets the server state to active.
func (srv *Server) start() error {
	srv.state.mu.Lock()
	defer srv.state.mu.Unlock()
	switch srv.state.value {
	case srvStateActive:
		return fmt.Errorf("go-kit: the server is already running")
	case srvStateStopped:
		return fmt.Errorf("go-kit: the server is in the stopped state. Waiting for shutdown. ")
	case srvStateClosed:
		return ErrServerClosed
	}
	srv.state.value = srvStateActive
	return nil
}

// Shutdown gracefully shuts down the server.
// It gracefully closes all active workers. The server will wait for
// active workers to finish processing tasks for duration specified in Config.ShutdownTimeout.
// If worker didn't finish processing a task during the timeout, the task will be pushed back to Redis.
func (srv *Server) Shutdown() {
	srv.state.mu.Lock()
	if srv.state.value == srvStateNew || srv.state.value == srvStateClosed {
		srv.state.mu.Unlock()
		// server is not running, do nothing and return.
		return
	}
	srv.state.value = srvStateClosed
	srv.state.mu.Unlock()

	// Note: The order of shutdown is important.
	// Sender goroutines should be terminated before the receiver goroutines.
	// processor -> syncer (via syncCh)
	// processor -> heartbeat (via starting, finished channels)
	srv.forwarder.shutdown()
	srv.processor.shutdown()
	srv.recoverer.shutdown()
	srv.syncer.shutdown()
	srv.subscriber.shutdown()
	srv.janitor.shutdown()
	srv.healthchecker.shutdown()
	srv.heartbeat.shutdown()
	srv.wg.Wait()

	srv.broker.Close()
}

// Stop signals the server to stop pulling new tasks off queues.
// Stop can be used before shutting down the server to ensure that all
// currently active tasks are processed before server shutdown.
//
// Stop does not shut down the server, make sure to call Shutdown before exit.
func (srv *Server) Stop() {
	srv.state.mu.Lock()
	if srv.state.value != srvStateActive {
		// Invalid call to Stop, server can only go from Active state to Stopped state.
		srv.state.mu.Unlock()
		return
	}
	srv.state.value = srvStateStopped
	srv.state.mu.Unlock()

	srv.processor.stop()
}
