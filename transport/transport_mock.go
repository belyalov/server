package transport

// MockTransport is dummy implementation of Transport interface
// mostly for unittests.
type MockTransport struct {
	BaseTransport

	Ch      chan []byte
	Error   error
	History [][]byte
}

// NewMockTransport creates mock transport
func NewMockTransport(name string) *MockTransport {
	return &MockTransport{
		BaseTransport: BaseTransport{
			Name: name,
		},
		Ch: make(chan []byte),
	}
}

// Start does nothing, it just returns preset error
func (m *MockTransport) Start() error {
	return m.Error
}

// Stop does nothing
func (m *MockTransport) Stop() {
}

// Receive returns pre created channel
func (m *MockTransport) Receive() <-chan []byte {
	return m.Ch
}

// Send does not actually send something, just adds
// packet into history
func (m *MockTransport) Send(packet []byte) error {
	m.History = append(m.History, packet)

	return nil
}
