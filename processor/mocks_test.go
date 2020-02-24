package processor

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
