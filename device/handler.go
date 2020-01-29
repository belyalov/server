package device


// Handler defines Device Handler - a way process device messages
type Handler interface {
	GetName() string
	ProcessMessage() error
	AddDevice(device *Device)
}
