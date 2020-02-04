package handlers

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/open-iot-devices/server/device"
	"github.com/stretchr/testify/assert"
)

func TestDeviceHandlers(t *testing.T) {
	// Add one
	handler := &mockHandler{}
	MustAddDeviceHandler(handler)
	// Add it one more time
	assert.Panics(t, func() {
		MustAddDeviceHandler(handler)
	})

	// Lookup it
	assert.NotNil(t, FindDeviceHandler(handler.GetName()))
	assert.Equal(t, map[string]device.Handler{"mockHandler": handler}, GetAllHandlers())
	// Lookup non existing device handler
	assert.Nil(t, FindDeviceHandler("test111fsdfsd"))
}

// Mocks
type mockHandler struct{}

func (*mockHandler) GetName() string {
	return "mockHandler"
}

func (*mockHandler) ProcessMessage(msg proto.Message) error {
	return nil
}

func (*mockHandler) AddDevice(device *device.Device) {

}
