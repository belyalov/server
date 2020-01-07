package transport

import (
	"context"

	"github.com/golang/protobuf/proto"

	"github.com/open-iot-devices/server/device"
)

// type Message struct {
// 	system   bool
// 	deviceID uint64
// }

// Transport represent a OpenIoT transport layer, e.g.
// UDP, TCP, USB, etc
type Transport interface {
	Run(context.Context) error
	Receive() <-chan []byte
	SendProtobuf(device device.Device, msg proto.Message) error
}
