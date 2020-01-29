package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	transport := &mockTransport{}
	MustAddTransport(transport)
	// Add it one more time
	assert.Panics(t, func() {
		MustAddTransport(transport)
	})

	// Lookup it
	assert.NotNil(t, FindTransportByName(transport.GetName()))
	// Lookup non existing device handler
	assert.Nil(t, FindTransportByName("test111fsdfsd"))
}

// Mock Transport
type mockTransport struct{}

func (m *mockTransport) GetName() string {
	return "name"
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
