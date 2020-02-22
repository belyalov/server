package transport

// Mock Transport
type mockTransport struct {
	Str string
	Int int

	name string
}

func newMockTransport(name string) Transport {
	return &mockTransport{
		name: name,
	}
}

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
