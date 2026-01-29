package http

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"time"
)

// ServerOption is an HTTP server option.
type ServerOption func(*Server)

// Address with server address.
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Timeout with server timeout.
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// Handler register router server .
func Handler(handle http.Handler) ServerOption {
	return func(s *Server) {
		s.Serve.Handler = handle
	}
}

// Alias set alias for server
func Alias(name string) ServerOption {
	return func(server *Server) {
		server.alias = name
	}
}

// Server is an HTTP server wrapper.
type Server struct {
	Serve   *http.Server
	lis     net.Listener
	err     error
	alias   string
	network string
	address string
	timeout time.Duration
}

// NewServer .
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		Serve:   new(http.Server),
		network: "tcp",
		address: ":0",
		timeout: 1 * time.Second,
	}

	for _, o := range opts {
		o(srv)
	}

	if srv.lis == nil {
		srv.lis, srv.err = net.Listen(srv.network, srv.address)
	}

	return srv
}

// Start . start the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	if s.err != nil {
		return s.err
	}
	s.Serve.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	err := s.Serve.Serve(s.lis)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop . stop the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	return s.Serve.Shutdown(ctx)
}

func (s *Server) Addr(b *bytes.Buffer) {
	if b.Len() > 0 {
		b.WriteString(", ")
	}

	if len(s.alias) > 0 {
		b.WriteString(s.alias + ":")
	}
	b.WriteString(s.lis.Addr().String())
}
