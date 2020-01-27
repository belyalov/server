package transport

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/open-iot-devices/server/device"
)

type MockTransport struct {
	Ch      chan []byte
	Error   error
	History [][]byte
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		Ch: make(chan []byte),
	}
}

func (m *MockTransport) Start(context.Context) error {
	return m.Error
}

func (m *MockTransport) Receive() <-chan []byte {
	return m.Ch
}

func (m *MockTransport) SendProtobuf(device device.Device, msg proto.Message) error {
	return nil
}
