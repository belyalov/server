package device

import "github.com/golang/protobuf/proto"

// Handler defines Device Handler - a way process device messages
type Handler interface {
	GetName() string
	ProcessMessage(msg proto.Message) error
	AddDevice(device *Device)
}
