package transport

import (
	"context"
)

// ServerInterface is transport server.
type ServerInterface interface {
	Start(context.Context) error
	Stop(context.Context) error
	//Addr(b *bytes.Buffer)
}
