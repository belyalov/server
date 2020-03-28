package device

import "github.com/golang/protobuf/proto"

type mockHandler struct {
	name    string
	history []*Device
}

type mockTransport struct {
	name string
}

// Mock Handler
func (m *mockHandler) GetName() string {
	return m.name
}

func (*mockHandler) ProcessMessage(device *Device, msg proto.Message) error {
	return nil
}

func (m *mockHandler) AddDevice(device *Device) {
	m.history = append(m.history, device)
}

// Mock Transport

func (m *mockTransport) GetName() string {
	return m.name
}

func (m *mockTransport) GetTypeName() string {
	return "mock"
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
