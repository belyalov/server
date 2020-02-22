package transport

// Transport represent a OpenIoT transport layer, e.g.
// UDP, TCP, USB, etc
type Transport interface {
	GetName() string
	GetTypeName() string
	Start() error
	Stop()

	Receive() <-chan []byte
	Send([]byte) error
}
