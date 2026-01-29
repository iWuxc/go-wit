package server

import "sync"

type serverState struct {
	mu    sync.Mutex
	value serverStateValue
}

type serverStateValue int

const (
	// StateNew represents a new server. Server begins in
	// this state and then transition to StatusActive when
	// Start or Run is callled.
	srvStateNew serverStateValue = iota

	// StateActive indicates the server is up and active.
	srvStateActive

	// StateStopped indicates the server is up but no longer processing new tasks.
	srvStateStopped

	// StateClosed indicates the server has been shutdown.
	srvStateClosed
)

var serverStates = []string{
	"new",
	"active",
	"stopped",
	"closed",
}

func (s serverStateValue) String() string {
	if srvStateNew <= s && s <= srvStateClosed {
		return serverStates[s]
	}
	return "unknown status"
}
