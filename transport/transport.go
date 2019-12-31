package transport

import (
	"context"
)

// Transport represent a OpenIoT transport layer, e.g.
// UDP, TCP, USB, etc
type Transport interface {
	Run(context.Context) error
	Receive() <-chan []byte
	Send([]byte) error
}
