package config

import (
	"github.com/golang/protobuf/proto"
	"github.com/open-iot-devices/server/device"
)

type mockHandler struct {
	name string
}

type mockTransport struct {
	name string
}

// Mock Handler
func (m *mockHandler) GetName() string {
	return m.name
}

func (*mockHandler) ProcessMessage(msg proto.Message) error {
	return nil
}

func (m *mockHandler) AddDevice(device *device.Device) {
}

// Mock Transport

func (m *mockTransport) GetName() string {
	return m.name
}

func (m *mockTransport) Start() error {
	return nil
}

func (m *mockTransport) Stop() {
}

func (m *mockTransport) Receive() <-chan []byte {
	return nil
}

func (m *mockTransport) Send([]byte) error {
	return nil
}
