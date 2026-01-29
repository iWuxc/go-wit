package app

import (
	"context"
	"github.com/iWuxc/go-wit/transport"
	"github.com/go-kratos/kratos/v2/registry"
	"net/url"
	"os"
	"time"
)

// Option is an application option.
type Option func(o *options, app *App)

// options is an application options.
type options struct {
	id               string
	name             string
	endpoints        []*url.URL
	ctx              context.Context
	sigs             []os.Signal
	stopTimeout      time.Duration
	servers          []transport.ServerInterface
	metadata         map[string]string
	version          string
	registrar        registry.Registrar
	registrarTimeout time.Duration
}

// ID with service id.
func ID(id string) Option {
	return func(o *options, app *App) { o.id = id }
}

// Name with service name.
func Name(name string) Option {
	return func(o *options, app *App) { o.name = name }
}

// Version with service version.
func Version(version string) Option {
	return func(o *options, app *App) { app.version = version }
}

// Context with service context.
func Context(ctx context.Context) Option {
	return func(o *options, app *App) { o.ctx = ctx }
}

// Server with transport servers.
func Server(srv ...transport.ServerInterface) Option {
	return func(o *options, app *App) { o.servers = srv }
}

// Signal with exit signals.
func Signal(sigs ...os.Signal) Option {
	return func(o *options, app *App) { o.sigs = sigs }
}

// StopTimeout with app stop timeout.
func StopTimeout(t time.Duration) Option {
	return func(o *options, app *App) { o.stopTimeout = t }
}

// Metadata with service metadata.
func Metadata(md map[string]string) Option {
	return func(o *options, app *App) {
		o.metadata = md
	}
}

// Registrar with service registry.
func Registrar(r registry.Registrar) Option {
	return func(o *options, app *App) { o.registrar = r }
}

// RegistrarTimeout with registrar timeout.
func RegistrarTimeout(t time.Duration) Option {
	return func(o *options, app *App) { o.registrarTimeout = t }
}

// Endpoint with service endpoint.
func Endpoint(endpoints ...*url.URL) Option {
	return func(o *options, app *App) { o.endpoints = endpoints }
}
