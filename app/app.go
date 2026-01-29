package app

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/metrics/stat"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
)

type App struct {
	mx       sync.Mutex
	opts     options
	ctx      context.Context
	cancel   func()
	version  string
	instance *registry.ServiceInstance
}

func NewApp(opts ...Option) (*App, error) {
	a := new(App)
	a.opts = options{
		ctx:         context.Background(),
		sigs:        []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		stopTimeout: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(&a.opts, a)
	}

	a.ctx, a.cancel = context.WithCancel(a.opts.ctx)
	return a, nil
}

func (a *App) Run() error {
	instance, err := a.buildInstance()
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(a.ctx)
	addr := strings.Join(instance.Endpoints, ":")
	wg := sync.WaitGroup{}
	for _, server := range a.opts.servers {
		srv := server
		//srv.Addr(addr)
		eg.Go(func() error {
			<-ctx.Done() // wait for stop signal
			sctx, cancel := context.WithTimeout(context.Background(), a.opts.stopTimeout)
			defer cancel()
			return srv.Stop(sctx)
		})
		wg.Add(1)
		eg.Go(func() error {
			wg.Done()
			return srv.Start(ctx)
		})
	}

	wg.Wait()
	if a.opts.registrar != nil {
		rctx, rcancel := context.WithTimeout(ctx, a.opts.registrarTimeout)
		defer rcancel()
		if err = a.opts.registrar.Register(rctx, instance); err != nil {
			return err
		}
	}

	pid := os.Getpid()
	stat.APPBootInfo.With("app_name", a.opts.name, "app_pid", fmt.Sprintf("%d", pid), "app_port", addr, "app_version", a.version, "kit_version", a.getKitVersion()).Set(1)
	stat.APPBootSeconds.Set(float64(time.Now().Unix()))
	log.Printf("Serving %s start with pid: %d and Version: %s", addr, pid, a.version)
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c:
				err := a.Stop()
				if err != nil {
					log.Errorf("failed to stop app: %v", err)
					return err
				}
				log.Printf("Serving %s has Done. ", addr)
			}
		}
	})
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

// Stop gracefully stops the application.
func (a *App) Stop() error {

	a.mx.Lock()
	instance := a.instance
	a.mx.Unlock()
	if a.opts.registrar != nil && instance != nil {
		ctx, cancel := context.WithTimeout(NewContext(a.ctx, a), a.opts.registrarTimeout)
		defer cancel()
		if err := a.opts.registrar.Deregister(ctx, instance); err != nil {
			return err
		}
	}

	if a.cancel != nil {
		a.cancel()
	}
	return nil
}

func (a *App) getKitVersion() string {
	res, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, dep := range res.Deps {
		if strings.HasPrefix(dep.Path, "git.microdreams.com") {
			return dep.Version
		}
	}

	return ""
}

func (a *App) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0, len(a.opts.endpoints))
	for _, e := range a.opts.endpoints {
		endpoints = append(endpoints, e.String())
	}
	if len(endpoints) == 0 {
		for _, srv := range a.opts.servers {
			if r, ok := srv.(transport.Endpointer); ok {
				e, err := r.Endpoint()
				if err != nil {
					return nil, err
				}
				endpoints = append(endpoints, e.String())
			}
		}
	}
	return &registry.ServiceInstance{
		ID:        a.opts.id,
		Name:      a.opts.name,
		Version:   a.version,
		Metadata:  a.opts.metadata,
		Endpoints: endpoints,
	}, nil
}

// AppInfo is application context value.
type AppInfo interface {
	ID() string
	Name() string
	Version() string
	Metadata() map[string]string
	Endpoint() []string
}

type appKey struct{}

// NewContext returns a new Context that carries value.
func NewContext(ctx context.Context, s AppInfo) context.Context {
	return context.WithValue(ctx, appKey{}, s)
}

// FromContext returns the Transport value stored in ctx, if any.
func FromContext(ctx context.Context) (s AppInfo, ok bool) {
	s, ok = ctx.Value(appKey{}).(AppInfo)
	return
}

// ID returns app instance id.
func (a *App) ID() string { return a.opts.id }

// Name returns service name.
func (a *App) Name() string { return a.opts.name }

// Version returns app version.
func (a *App) Version() string { return a.opts.version }

// Metadata returns service metadata.
func (a *App) Metadata() map[string]string { return a.opts.metadata }

// Endpoint returns endpoints.
func (a *App) Endpoint() []string {
	if a.instance != nil {
		return a.instance.Endpoints
	}
	return nil
}
