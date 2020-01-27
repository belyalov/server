package transport

import (
	"context"
)

// Transport represent a OpenIoT transport layer, e.g.
// UDP, TCP, USB, etc
type Transport interface {
	GetName() string
	Start(context.Context) error

	Receive() <-chan []byte
	Send([]byte) error
}

// BaseTransport implements basic methods
type BaseTransport struct {
	Name string
}

// GetName returns name of transport
func (b *BaseTransport) GetName() string {
	return b.Name
}
