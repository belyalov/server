package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceHandler(t *testing.T) {
	handler := &mockHandler{}

	dev := NewDevice(1111)
	err := dev.AddDeviceHandler(handler)
	assert.NoError(t, err)

	// Add the same handler
	err = dev.AddDeviceHandler(handler)
	assert.Error(t, err)

	// Device handler's array is correct
	assert.Equal(t, []Handler{handler}, dev.Handlers())

	// Ensure that handler.AddDevice is actually called
	assert.Equal(t, []*Device{dev}, handler.history)
}

func TestDeviceTransport(t *testing.T) {
	dev := NewDevice(123)
	transport := &mockTransport{}

	dev.SetTransport(transport)

	assert.Equal(t, transport, dev.transport)
	assert.Equal(t, transport.GetName(), dev.TransportName)
	assert.Equal(t, transport, dev.Transport())
}
