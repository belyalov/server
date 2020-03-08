package processor

import "bytes"

type mockTransport struct {
	history [][]byte
}

func (m *mockTransport) GetName() string {
	return "mock"
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

func (m *mockTransport) Send(msg []byte) error {
	m.history = append(m.history, msg)

	return nil
}

func (m *mockTransport) Empty() bool {
	return len(m.history) == 0
}

func (m *mockTransport) LastMessage() *bytes.Buffer {
	size := len(m.history)
	if size == 0 {
		panic("mockTransport history is empty")
	}
	return bytes.NewBuffer(m.history[size-1])
}
