//go:build linux || bsd || darwin
// +build linux bsd darwin

package server

import (
	"github.com/iWuxc/go-wit/log"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

// waitForSignals waits for signals and handles them.
// It handles SIGTERM, SIGINT, and SIGTSTP.
// SIGTERM and SIGINT will signal the process to exit.
// SIGTSTP will signal the process to stop processing new tasks.
func (srv *Server) waitForSignals() {
	log.Info("Send signal TSP to stop processing new tasks")
	log.Info("Send signal TERM or INT to terminate the process")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT, unix.SIGTSTP)
	for {
		sig := <-sigs
		if sig == unix.SIGTSTP {
			srv.Stop()
			continue
		}
		break
	}
}
